package plugin

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j/dbtype"
)

func ptrS(v string) *string       { return &v }
func ptrI(v int64) *int64         { return &v }
func ptrF(v float64) *float64     { return &v }
func ptrB(v bool) *bool           { return &v }
func ptrT(v time.Time) *time.Time { return &v }

// diffFrames compares two frames the way the SDK's own tests do.
func diffFrames(got, want *data.Frame) string {
	return cmp.Diff(got, want, data.FrameTestCompareOptions()...)
}

// runIntegrationQuery runs one cypher query against the local compose Neo4j.
func runIntegrationQuery(t *testing.T, cypher, format string) backend.DataResponse {
	t.Helper()
	settings := settingsFor("neo4j://localhost:7687", "", "neo4j", "Password123")
	inst, err := NewDatasource(context.Background(), settings)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { inst.(*Datasource).Dispose() })
	ds := inst.(*Datasource)
	raw, _ := json.Marshal(QueryModel{CypherQuery: cypher, Format: format})
	return ds.query(context.Background(), backend.DataQuery{RefID: "A", JSON: raw})
}

func TestTableFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	_ = dbtype.Node{} // keep the dbtype import honest for future graph tests
	cases := []struct {
		name     string
		cypher   string
		expected *data.Frame
	}{
		{"string", "With 'One' as A return A",
			data.NewFrame("response", data.NewField("A", nil, []*string{ptrS("One")}))},
		{"int", "With 1 as A return A",
			data.NewFrame("response", data.NewField("A", nil, []*int64{ptrI(1)}))},
		{"float", "With 0.81234 as A return A",
			data.NewFrame("response", data.NewField("A", nil, []*float64{ptrF(0.81234)}))},
		{"bool", "With TRUE as A return A UNION ALL With FALSE as A return A",
			data.NewFrame("response", data.NewField("A", nil, []*bool{ptrB(true), ptrB(false)}))},
		{"multiple rows and columns", "With 'One' as A, 'Two' as B, 1 as C return A, B, C UNION ALL With 'Three' as A, 'Four' as B, 2 as C return A, B, C",
			data.NewFrame("response",
				data.NewField("A", nil, []*string{ptrS("One"), ptrS("Three")}),
				data.NewField("B", nil, []*string{ptrS("Two"), ptrS("Four")}),
				data.NewField("C", nil, []*int64{ptrI(1), ptrI(2)}))},
		{"null in typed column", "With null as A return A UNION ALL With 1 as A return A",
			data.NewFrame("response", data.NewField("A", nil, []*int64{nil, ptrI(1)}))},
		{"all null column falls back to string", "With null as A return A UNION ALL With null as A return A",
			data.NewFrame("response", data.NewField("A", nil, []*string{nil, nil}))},
		{"list as json", "Return ['a', 'b', 'c'] as A",
			data.NewFrame("response", data.NewField("A", nil, []*string{ptrS("[\"a\",\"b\",\"c\"]")}))},
		{"map as json", "RETURN {key: 'Value'} as A",
			data.NewFrame("response", data.NewField("A", nil, []*string{ptrS("{\"key\":\"Value\"}")}))},
		{"datetime", "return datetime(\"2022-03-02T13:14:15.144Z\") as A",
			data.NewFrame("response", data.NewField("A", nil, []*time.Time{ptrT(time.Date(2022, 3, 2, 13, 14, 15, 144000000, time.UTC))}))},
		{"duration as string", "return duration(\"PT3M\") as A",
			data.NewFrame("response", data.NewField("A", nil, []*string{ptrS("PT3M")}))},
		{"node as json", "Match(m:Movie) where m.title = 'The Matrix' return m limit 1",
			data.NewFrame("response", data.NewField("m", nil, []*string{ptrS("{\"Id\":0,\"ElementId\":\"4:0e328ed8-939c-4e4f-ae9a-3fdec9079c6e:0\",\"Labels\":[\"Movie\"],\"Props\":{\"released\":1999,\"tagline\":\"Welcome to the Real World\",\"title\":\"The Matrix\"}}")}))},
		{"relationship as json", "MATCH (p:Person)-[r:ACTED_IN]->(m:Movie) where m.title = 'The Matrix' AND p.name = 'Keanu Reeves' RETURN r LIMIT 1",
			data.NewFrame("response", data.NewField("r", nil, []*string{ptrS("{\"Id\":0,\"ElementId\":\"5:0e328ed8-939c-4e4f-ae9a-3fdec9079c6e:0\",\"StartId\":1,\"StartElementId\":\"4:0e328ed8-939c-4e4f-ae9a-3fdec9079c6e:1\",\"EndId\":0,\"EndElementId\":\"4:0e328ed8-939c-4e4f-ae9a-3fdec9079c6e:0\",\"Type\":\"ACTED_IN\",\"Props\":{\"roles\":[\"Neo\"]}}")}))},
		{"dotted column name", "Match(m:Movie) where m.title = 'The Matrix' return m.title limit 1",
			data.NewFrame("response", data.NewField("m.title", nil, []*string{ptrS("The Matrix")}))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := runIntegrationQuery(t, tc.cypher, "table")
			if len(res.Frames) != 1 {
				t.Fatalf("expected 1 frame, got %d (error: %v)", len(res.Frames), res.Error)
			}
			if diff := diffFrames(res.Frames[0], tc.expected); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

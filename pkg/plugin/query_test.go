package plugin

import (
	"context"
	"encoding/json"
	"fmt"
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

// newTestDatasource builds a *Datasource against the local compose Neo4j and
// disposes it when the test ends.
func newTestDatasource(t *testing.T) *Datasource {
	t.Helper()
	settings := settingsFor("neo4j://localhost:7687", "", "neo4j", "Password123")
	inst, err := NewDatasource(context.Background(), settings)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { inst.(*Datasource).Dispose() })
	return inst.(*Datasource)
}

// runIntegrationQuery runs one cypher query against the local compose Neo4j.
func runIntegrationQuery(t *testing.T, cypher, format string) backend.DataResponse {
	t.Helper()
	ds := newTestDatasource(t)
	raw, _ := json.Marshal(QueryModel{CypherQuery: cypher, Format: format})
	return ds.query(context.Background(), backend.DataQuery{RefID: "A", JSON: raw})
}

// probeNodeRelIds returns the internal and element ids of the
// Keanu Reeves --ACTED_IN--> The Matrix pattern from the dev seed. Both id
// kinds depend on the database instance (element ids embed the database
// UUID, internal ids depend on creation order), so tests must never
// hard-code them; build expectations from the probed values instead.
func probeNodeRelIds(t *testing.T) (pid int64, peid string, rid int64, reid string, mid int64, meid string) {
	t.Helper()
	res := runIntegrationQuery(t,
		"MATCH (p:Person {name:'Keanu Reeves'})-[r:ACTED_IN]->(m:Movie {title:'The Matrix'}) "+
			"RETURN id(p) AS pid, elementId(p) AS peid, id(r) AS rid, elementId(r) AS reid, id(m) AS mid, elementId(m) AS meid",
		formatTable)
	if len(res.Frames) != 1 || res.Frames[0].Rows() != 1 {
		t.Fatalf("id probe: expected 1 frame with 1 row, got %d frames (error: %v)", len(res.Frames), res.Error)
	}
	frame := res.Frames[0]
	field := func(name string) any {
		for _, f := range frame.Fields {
			if f.Name == name {
				return f.At(0)
			}
		}
		t.Fatalf("id probe: no field %q", name)
		return nil
	}
	asInt := func(name string) int64 {
		v, _ := field(name).(*int64)
		if v == nil {
			t.Fatalf("id probe: %q is not an integer: %v", name, field(name))
		}
		return *v
	}
	asStr := func(name string) string {
		v, _ := field(name).(*string)
		if v == nil {
			t.Fatalf("id probe: %q is not a string: %v", name, field(name))
		}
		return *v
	}
	return asInt("pid"), asStr("peid"), asInt("rid"), asStr("reid"), asInt("mid"), asStr("meid")
}

func TestTableFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	_ = dbtype.Node{} // keep the dbtype import honest for future graph tests
	pid, peid, rid, reid, mid, meid := probeNodeRelIds(t)
	nodeJSON := fmt.Sprintf(`{"Id":%d,"ElementId":"%s","Labels":["Movie"],"Props":{"released":1999,"tagline":"Welcome to the Real World","title":"The Matrix"}}`, mid, meid)
	relJSON := fmt.Sprintf(`{"Id":%d,"ElementId":"%s","StartId":%d,"StartElementId":"%s","EndId":%d,"EndElementId":"%s","Type":"ACTED_IN","Props":{"roles":["Neo"]}}`, rid, reid, pid, peid, mid, meid)
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
			data.NewFrame("response", data.NewField("m", nil, []*string{ptrS(nodeJSON)}))},
		{"relationship as json", "MATCH (p:Person)-[r:ACTED_IN]->(m:Movie) where m.title = 'The Matrix' AND p.name = 'Keanu Reeves' RETURN r LIMIT 1",
			data.NewFrame("response", data.NewField("r", nil, []*string{ptrS(relJSON)}))},
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

func TestQueryDatabaseOverride(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	// Community Edition has only the default "neo4j" database besides
	// "system", and the system database rejects user-database Cypher such as
	// RETURN 1 (Neo.ClientError.Statement.SemanticError), so the override is
	// proven by contrast instead: a per-query database must win over both a
	// broken and a working configured database.

	// (a) The per-query database rescues a query whose data source is
	// configured against a database that does not exist.
	settings := settingsFor("neo4j://localhost:7687", "DoesNotExist", "neo4j", "Password123")
	inst, err := NewDatasource(context.Background(), settings)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { inst.(*Datasource).Dispose() })
	poisoned := inst.(*Datasource)
	rawGood, _ := json.Marshal(QueryModel{CypherQuery: "RETURN 1 AS ok", Format: formatTable, Database: "neo4j"})
	if res := poisoned.query(context.Background(), backend.DataQuery{RefID: "A", JSON: rawGood}); res.Error != nil {
		t.Fatalf("per-query database override failed: %v", res.Error)
	}

	// (b) A per-query database pointing at a database that does not exist
	// fails even though the data source's default would have worked.
	ds := newTestDatasource(t)
	rawBad, _ := json.Marshal(QueryModel{CypherQuery: "RETURN 1 AS ok", Database: "DoesNotExist"})
	if res := ds.query(context.Background(), backend.DataQuery{RefID: "A", JSON: rawBad}); res.Error == nil {
		t.Fatal("expected error for non-existent database")
	}
}

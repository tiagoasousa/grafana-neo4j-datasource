package plugin

import (
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// probeElementIds runs the elementId probe query through the datasource's
// normal query pipeline (table format) and returns the person, relationship
// and movie element ids of the Keanu Reeves --ACTED_IN--> The Matrix pattern.
// On Neo4j 5, element ids look like `4:<db-uuid>:<id>` where the database UUID
// is tied to the compose data volume, so tests must never hard-code them.
func probeElementIds(t *testing.T) (person, rel, movie string) {
	t.Helper()
	res := runIntegrationQuery(t,
		"MATCH (p:Person {name:'Keanu Reeves'})-[r:ACTED_IN]->(m:Movie {title:'The Matrix'}) "+
			"RETURN elementId(p) AS pe, elementId(r) AS re, elementId(m) AS me",
		formatTable)
	if len(res.Frames) != 1 || len(res.Frames[0].Fields) != 3 {
		t.Fatalf("elementId probe: expected 1 frame with 3 fields, got %d frames (error: %v)", len(res.Frames), res.Error)
	}
	frame := res.Frames[0]
	if frame.Rows() != 1 {
		t.Fatalf("elementId probe: expected 1 row, got %d", frame.Rows())
	}
	row := map[string]string{}
	for _, f := range frame.Fields {
		v, _ := f.At(0).(*string)
		if v == nil {
			t.Fatalf("elementId probe: field %q is not a string: %v", f.Name, f.At(0))
		}
		row[f.Name] = *v
	}
	return row["pe"], row["re"], row["me"]
}

func TestGraphFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	skipIfNoNeo4j(t)
	personEid, relEid, movieEid := probeElementIds(t)
	t.Logf("probed element ids: person=%q rel=%q movie=%q", personEid, relEid, movieEid)

	cypher := "Match(m:Movie)<-[a:ACTED_IN]-(f:Person) where m.title = 'The Matrix' and f.name = 'Keanu Reeves' return m.title, f, a, m order by m.title"

	res := runIntegrationQuery(t, cypher, formatNodeGraph)
	if len(res.Frames) != 2 {
		t.Fatalf("expected 2 frames, got %d (error: %v)", len(res.Frames), res.Error)
	}

	expectedNodes := data.NewFrame("nodes",
		data.NewField("id", nil, []*string{ptrS(personEid), ptrS(movieEid)}),
		data.NewField("title", nil, []*string{ptrS("Person"), ptrS("The Matrix")}),
		data.NewField("detail__labels", nil, []*string{nil, nil}),
		data.NewField("detail__born", nil, []*string{ptrS("1964"), nil}),
		data.NewField("detail__name", nil, []*string{ptrS("Keanu Reeves"), nil}),
		data.NewField("detail__released", nil, []*string{nil, ptrS("1999")}),
		data.NewField("detail__tagline", nil, []*string{nil, ptrS("Welcome to the Real World")}),
	)
	expectedEdges := data.NewFrame("edges",
		data.NewField("id", nil, []*string{ptrS(relEid)}),
		data.NewField("source", nil, []*string{ptrS(personEid)}),
		data.NewField("target", nil, []*string{ptrS(movieEid)}),
		data.NewField("mainStat", nil, []*string{ptrS("ACTED_IN")}),
		data.NewField("detail__roles", nil, []*string{ptrS("[\"Neo\"]")}),
	)
	m := data.FrameMeta{PreferredVisualization: "nodeGraph"}
	expectedNodes = expectedNodes.SetMeta(&m)
	expectedEdges = expectedEdges.SetMeta(&m)

	if diff := diffFrames(res.Frames[0], expectedNodes); diff != "" {
		t.Fatal("nodes:", diff)
	}
	if diff := diffFrames(res.Frames[1], expectedEdges); diff != "" {
		t.Fatal("edges:", diff)
	}
}

func TestGraphFormatOptionalFields(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	skipIfNoNeo4j(t)
	cypher := `
MERGE (n:OptionalFieldTest {name: 'X'})
SET n.subtitle = 'sub', n.color = 'red', n.mainstat = '42', n.arc__cpu = 5, n.other = 'plain'
RETURN n`

	res := runIntegrationQuery(t, cypher, formatNodeGraph)
	if len(res.Frames) < 1 {
		t.Fatalf("expected a nodes frame, got %d frames (error: %v)", len(res.Frames), res.Error)
	}
	nodesFrame := res.Frames[0]
	row := map[string]string{}
	for _, f := range nodesFrame.Fields {
		v, _ := f.At(0).(*string)
		if v == nil {
			row[f.Name] = ""
		} else {
			row[f.Name] = *v
		}
	}
	expected := map[string]string{
		"title":         "OptionalFieldTest", // no title prop -> first label
		"subtitle":      "sub",
		"color":         "red",
		"mainstat":      "42",
		"arc__cpu":      "5",
		"detail__name":  "X",
		"detail__other": "plain",
	}
	for name, want := range expected {
		if got := row[name]; got != want {
			t.Errorf("field %q: want %q, got %q", name, want, got)
		}
	}
	if _, exists := row["detail__color"]; exists {
		t.Error("color must not be emitted as detail__color")
	}
}

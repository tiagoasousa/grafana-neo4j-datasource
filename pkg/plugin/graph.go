package plugin

import (
	"sort"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j/dbtype"
)

// isNodeGraphOptionalField reports whether a node/relationship property keeps
// its documented Node Graph field name instead of the detail__ prefix.
// https://grafana.com/docs/grafana/latest/panels-visualizations/visualizations/node-graph/
func isNodeGraphOptionalField(isNode bool, name string) bool {
	if strings.HasPrefix(name, "arc__") {
		return isNode // arc segments are node-only
	}
	if isNode {
		switch name {
		case "subtitle", "mainstat", "secondarystat", "color", "icon", "nodeRadius", "highlighted":
			return true
		}
		return false
	}
	switch name {
	case "secondarystat", "thickness", "color", "strokeDasharray", "highlighted":
		return true
	}
	return false
}

// toGraphResponse builds nodes/edges frames for the Node Graph panel.
func toGraphResponse(keys []string, records []*neo4j.Record) backend.DataResponse {
	response := backend.DataResponse{}

	nodesFrame, nodesPropIndex := newGraphFrame("nodes", []string{"id", "title", "detail__labels"}, records, true)
	edgesFrame, edgesPropIndex := newGraphFrame("edges", []string{"id", "source", "target", "mainStat"}, records, false)

	nodeIdMap := map[string]string{}
	for _, record := range records {
		for _, v := range record.Values {
			if node, ok := v.(dbtype.Node); ok {
				if _, seen := nodeIdMap[node.ElementId]; seen {
					continue
				}
				nodeIdMap[node.ElementId] = ""

				title := ""
				if len(node.Labels) > 0 {
					title = node.Labels[0]
				}
				if t, ok := node.Props["title"].(string); ok && t != "" {
					title = t
				}
				row := make([]any, len(nodesFrame.Fields))
				row[0] = &node.ElementId
				row[1] = &title
				if len(node.Labels) > 1 {
					s := asJson(node.Labels)
					row[2] = s
				}
				for name, value := range node.Props {
					if name == "title" {
						continue
					}
					row[nodesPropIndex[name]] = asJson(value)
				}
				nodesFrame.AppendRow(row...)
			}
		}
	}
	for _, record := range records {
		for _, v := range record.Values {
			edge, ok := v.(dbtype.Relationship)
			if !ok {
				continue
			}
			_, startExists := nodeIdMap[edge.StartElementId]
			_, endExists := nodeIdMap[edge.EndElementId]
			if !startExists || !endExists {
				continue
			}
			row := make([]any, len(edgesFrame.Fields))
			row[0] = &edge.ElementId
			row[1] = &edge.StartElementId
			row[2] = &edge.EndElementId
			row[3] = &edge.Type
			for name, value := range edge.Props {
				row[edgesPropIndex[name]] = asJson(value)
			}
			edgesFrame.AppendRow(row...)
		}
	}

	m := data.FrameMeta{PreferredVisualization: "nodeGraph"}
	response.Frames = append(response.Frames,
		nodesFrame.SetMeta(&m),
		edgesFrame.SetMeta(&m),
	)
	return response
}

// newGraphFrame creates the frame with meta fields plus one field per
// discovered property (optional fields unprefixed).
func newGraphFrame(name string, metaFields []string, records []*neo4j.Record, isNode bool) (*data.Frame, map[string]int) {
	fields := make([]*data.Field, 0, len(metaFields))
	for _, f := range metaFields {
		fields = append(fields, data.NewField(f, nil, []*string{}))
	}
	propSet := map[string]struct{}{}
	for _, record := range records {
		for _, v := range record.Values {
			// Collect props only from values matching this frame's type:
			// node props for the nodes frame, relationship props for edges.
			var props map[string]any
			if isNode {
				if n, ok := v.(dbtype.Node); ok {
					props = n.Props
				}
			} else if r, ok := v.(dbtype.Relationship); ok {
				props = r.Props
			}
			for p := range props {
				if isNode && p == "title" {
					continue // feeds the title meta field instead
				}
				propSet[p] = struct{}{}
			}
		}
	}
	names := make([]string, 0, len(propSet))
	for p := range propSet {
		names = append(names, p)
	}
	sort.Strings(names)

	propIndex := map[string]int{}
	idx := len(metaFields)
	for _, p := range names {
		if isNodeGraphOptionalField(isNode, p) {
			fields = append(fields, data.NewField(p, nil, []*string{}))
		} else {
			fields = append(fields, data.NewField("detail__"+p, nil, []*string{}))
		}
		propIndex[p] = idx
		idx++
	}
	return data.NewFrame(name, fields...), propIndex
}

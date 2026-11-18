package plugin

import (
	"encoding/json"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j/dbtype"
)

// toTableResponse builds a single frame, inferring each column's field type
// from the first non-nil value (falling back to string).
func toTableResponse(keys []string, records []*neo4j.Record) backend.DataResponse {
	response := backend.DataResponse{}
	frame := data.NewFrame("response")

	for col, name := range keys {
		typ := inferFieldType(records, col)
		frame.Fields = append(frame.Fields, data.NewField(name, nil, typ))
	}
	for _, record := range records {
		row := make([]any, len(frame.Fields))
		for col, value := range record.Values {
			row[col] = toValue(value)
		}
		frame.AppendRow(row...)
	}
	response.Frames = append(response.Frames, frame)
	return response
}

// inferFieldType returns a typed nil-slice for the first non-nil value in the
// column, defaulting to []*string when all values are nil.
func inferFieldType(records []*neo4j.Record, col int) any {
	for _, record := range records {
		if record.Values[col] == nil {
			continue
		}
		switch record.Values[col].(type) {
		case int64:
			return []*int64{}
		case float64:
			return []*float64{}
		case bool:
			return []*bool{}
		case time.Time, dbtype.Date, dbtype.Time, dbtype.LocalTime, dbtype.LocalDateTime:
			return []*time.Time{}
		default:
			return []*string{}
		}
	}
	return []*string{}
}

// toValue converts a driver value into a frame cell value.
func toValue(val any) any {
	switch t := val.(type) {
	case nil:
		return nil
	case string:
		return &t
	case int64:
		return &t
	case bool:
		return &t
	case float64:
		return &t
	case time.Time:
		return &t
	case dbtype.Date:
		v := t.Time()
		return &v
	case dbtype.Time:
		v := t.Time()
		return &v
	case dbtype.LocalTime:
		v := t.Time()
		return &v
	case dbtype.LocalDateTime:
		v := t.Time()
		return &v
	case dbtype.Duration:
		v := t.String()
		return &v
	default:
		return asJson(val)
	}
}

// asJson serializes complex values (nodes, relationships, maps, lists);
// plain strings pass through unquoted.
func asJson(val any) *string {
	if s, ok := val.(string); ok {
		return &s
	}
	r, err := json.Marshal(val)
	if err != nil {
		log.DefaultLogger.Error("json marshalling failed", "err", err.Error())
	}
	res := string(r)
	return &res
}

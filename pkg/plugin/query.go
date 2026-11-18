package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

const formatTable = "table"
const formatNodeGraph = "nodegraph"

// QueryModel is the query editor's persisted query.
type QueryModel struct {
	CypherQuery string `json:"cypherQuery"`
	Format      string `json:"format"`
	// Database optionally overrides the data source database for this query.
	Database string `json:"database"`
}

// runQuery executes one Cypher statement in an auto-commit transaction.
func (d *Datasource) runQuery(ctx context.Context, cypher, database string) ([]*neo4j.Record, []string, error) {
	db := d.settings.Database
	if database != "" {
		db = database
	}
	// Auto-commit user Cypher cannot be classified read/write up front, and
	// this data source intentionally allows writes (users may run MERGE/SET),
	// so sessions use write mode, which accepts reads as well.
	session := d.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: db, AccessMode: neo4j.AccessModeWrite})
	defer func() { _ = session.Close(ctx) }()

	result, err := session.Run(ctx, cypher, map[string]any{})
	if err != nil {
		return nil, nil, fmt.Errorf("could not run query: %w", err)
	}
	keys, err := result.Keys()
	if err != nil {
		return nil, nil, fmt.Errorf("could not read query keys: %w", err)
	}
	records, err := neo4j.Collect(ctx, result, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("could not collect query results: %w", err)
	}
	return records, keys, nil
}

func (d *Datasource) query(ctx context.Context, q backend.DataQuery) backend.DataResponse {
	var qm QueryModel
	if err := json.Unmarshal(q.JSON, &qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err))
	}
	records, keys, err := d.runQuery(ctx, qm.CypherQuery, qm.Database)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, err.Error())
	}
	if qm.Format == formatNodeGraph {
		return toGraphResponse(keys, records)
	}
	return toTableResponse(keys, records)
}

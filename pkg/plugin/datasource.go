package plugin

import (
	"context"
	"io"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/tiagoasousa/neo4j-datasource/pkg/models"
)

// Make sure Datasource implements required interfaces.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// Datasource answers Cypher queries against a Neo4j database over Bolt.
type Datasource struct {
	settings *models.PluginSettings
	driver   neo4j.DriverWithContext
	// relay is non-nil only when the secure SOCKS proxy routes Bolt
	// through the loopback relay (introduced in a later task).
	relay io.Closer
}

// NewDatasource creates a new datasource instance.
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	s, err := models.LoadPluginSettings(settings)
	if err != nil {
		return nil, err
	}
	driver, err := createDriver(s)
	if err != nil {
		return nil, err
	}
	return &Datasource{settings: s, driver: driver}, nil
}

// Dispose cleans up when settings change.
func (d *Datasource) Dispose() {
	if d.driver != nil {
		_ = d.driver.Close(context.Background())
	}
	if d.relay != nil {
		_ = d.relay.Close()
	}
}

// QueryData handles multiple queries and returns multiple responses.
// The query pipeline is not implemented yet; each query gets a placeholder
// error response until the next task wires it up.
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()
	for _, q := range req.Queries {
		response.Responses[q.RefID] = backend.ErrDataResponse(backend.StatusInternal, "query pipeline not implemented yet")
	}
	return response, nil
}

// CheckHealth verifies the driver can reach the server and that the
// configured database exists (the full query pipeline arrives in the next task).
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if err := d.driver.VerifyConnectivity(ctx); err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}
	// VerifyConnectivity only exercises the server's default database; probe
	// the configured database so a bad name fails the health check.
	session := d.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: d.settings.Database, AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)
	if _, err := session.Run(ctx, "RETURN 1", nil); err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}
	return &backend.CheckHealthResult{Status: backend.HealthStatusOk, Message: "Data source is working"}, nil
}

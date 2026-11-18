package plugin

import (
	"context"

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
	driver   neo4j.Driver
	// relay is non-nil only when the secure SOCKS proxy routes Bolt
	// through the loopback relay.
	relay *socksRelay
}

// NewDatasource creates a new datasource instance.
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	s, err := models.LoadPluginSettings(settings)
	if err != nil {
		return nil, err
	}
	httpOpts, err := settings.HTTPClientOptions(ctx)
	if err != nil {
		return nil, err
	}
	driver, relay, err := createDriver(ctx, s, httpOpts.ProxyOptions)
	if err != nil {
		return nil, err
	}
	return &Datasource{settings: s, driver: driver, relay: relay}, nil
}

// Dispose cleans up when settings change.
func (d *Datasource) Dispose() {
	if d.driver != nil {
		_ = d.driver.Close(context.Background())
	}
	if d.relay != nil {
		d.relay.close()
	}
}

// QueryData handles multiple queries and returns multiple responses.
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()
	for _, q := range req.Queries {
		response.Responses[q.RefID] = d.query(ctx, q)
	}
	return response, nil
}

// CheckHealth verifies the driver can reach the server and that the
// configured database exists.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if err := d.driver.VerifyConnectivity(ctx); err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}
	// VerifyConnectivity only exercises the server's default database; run a
	// real query against the configured database so a bad name fails health.
	if _, _, err := d.runQuery(ctx, "RETURN 1", ""); err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}
	return &backend.CheckHealthResult{Status: backend.HealthStatusOk, Message: "Data source is working"}, nil
}

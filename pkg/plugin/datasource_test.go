package plugin

import (
	"context"
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/tiagoasousa/neo4j-datasource/pkg/models"
)

func settingsFor(url, database, username, password string) backend.DataSourceInstanceSettings {
	s := models.PluginSettings{URL: url, Database: database, Username: username}
	s.Secrets = &models.SecretPluginSettings{Password: password}
	raw, _ := json.Marshal(s)
	return backend.DataSourceInstanceSettings{
		JSONData:                raw,
		DecryptedSecureJSONData: map[string]string{"password": password},
	}
}

// skipIfNoNeo4j skips integration tests when no Neo4j is reachable (e.g. CI
// release builds run the full suite without a database).
func skipIfNoNeo4j(t *testing.T) {
	t.Helper()
	conn, err := net.DialTimeout("tcp", "localhost:7687", 2*time.Second)
	if err != nil {
		t.Skip("Neo4j not reachable at localhost:7687; integration test skipped")
	}
	_ = conn.Close()
}

func TestHealthcheck(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	skipIfNoNeo4j(t)
	cases := []struct {
		name        string
		settings    backend.DataSourceInstanceSettings
		wantStatus  backend.HealthStatus
		wantMessage string
	}{
		{"ok", settingsFor("neo4j://localhost:7687", "", "neo4j", "Password123"), backend.HealthStatusOk, ""},
		{"invalid host", settingsFor("neo4j://invalid:7687", "", "neo4j", "Password123"), backend.HealthStatusError, ""},
		{"bad credentials", settingsFor("neo4j://localhost:7687", "", "doesNotExist", "Password123"), backend.HealthStatusError, "auth"},
		{"bad database", settingsFor("neo4j://localhost:7687", "DoesNotExist", "neo4j", "Password123"), backend.HealthStatusError, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inst, err := NewDatasource(context.Background(), tc.settings)
			if err != nil {
				t.Fatal(err)
			}
			ds := inst.(*Datasource)
			defer ds.Dispose()
			res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{
				PluginContext: backend.PluginContext{DataSourceInstanceSettings: &tc.settings},
			})
			if err != nil {
				t.Fatal(err)
			}
			if res.Status != tc.wantStatus {
				t.Errorf("status: want %v, got %v (%s)", tc.wantStatus, res.Status, res.Message)
			}
			if tc.wantMessage != "" && !strings.Contains(res.Message, tc.wantMessage) {
				t.Errorf("message %q should contain %q", res.Message, tc.wantMessage)
			}
		})
	}
}

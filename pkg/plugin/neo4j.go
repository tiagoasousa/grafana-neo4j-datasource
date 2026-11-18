package plugin

import (
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/tiagoasousa/neo4j-datasource/pkg/models"
)

func authTokenFor(s *models.PluginSettings) neo4j.AuthToken {
	if s.Username != "" && s.Secrets.Password != "" {
		return neo4j.BasicAuth(s.Username, s.Secrets.Password, "")
	}
	return neo4j.NoAuth()
}

// createDriver builds the Bolt driver (direct mode; PDC relay added in a later task).
func createDriver(s *models.PluginSettings) (neo4j.DriverWithContext, error) {
	driver, err := neo4j.NewDriverWithContext(s.URL, authTokenFor(s))
	if err != nil {
		log.DefaultLogger.Error("could not create neo4j driver", "url", s.URL, "err", err.Error())
		return nil, fmt.Errorf("could not create neo4j driver for %s: %w", s.URL, err)
	}
	return driver, nil
}

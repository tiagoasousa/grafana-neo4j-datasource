package plugin

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/proxy"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/tiagoasousa/neo4j-datasource/pkg/models"
)

func authTokenFor(s *models.PluginSettings) neo4j.AuthToken {
	if s.Username != "" && s.Secrets.Password != "" {
		return neo4j.BasicAuth(s.Username, s.Secrets.Password, "")
	}
	return neo4j.NoAuth()
}

// createDriver builds the Bolt driver. When the secure SOCKS proxy (Grafana
// Private Data Source Connect) is enabled for this data source, the driver
// connects to a local relay that tunnels every connection through the SOCKS
// proxy to the configured Neo4j address.
func createDriver(ctx context.Context, s *models.PluginSettings, proxyOpts *proxy.Options) (neo4j.Driver, *socksRelay, error) {
	proxyClient := proxy.New(proxyOpts)
	if proxyClient.SecureSocksProxyEnabled() {
		target, err := boltSocksTarget(s.URL)
		if err != nil {
			return nil, nil, err
		}
		dialer, err := proxyClient.NewSecureSocksProxyContextDialer()
		if err != nil {
			return nil, nil, fmt.Errorf("could not create secure socks dialer: %w", err)
		}
		relay, err := newSocksRelay(dialer, target)
		if err != nil {
			return nil, nil, fmt.Errorf("could not start local socks relay: %w", err)
		}
		log.DefaultLogger.Info("routing Bolt through secure SOCKS proxy", "target", target)
		driver, err := neo4j.NewDriver("bolt://"+relay.addr(), authTokenFor(s))
		return driver, relay, err
	}
	driver, err := neo4j.NewDriver(s.URL, authTokenFor(s))
	return driver, nil, err
}

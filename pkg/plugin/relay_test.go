package plugin

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/tiagoasousa/neo4j-datasource/pkg/models"
)

func TestBoltSocksTarget(t *testing.T) {
	cases := []struct {
		url      string
		expected string
		wantErr  bool
	}{
		{url: "neo4j://localhost:7687", expected: "localhost:7687"},
		{url: "bolt://neo4j.cloud-inventory.svc.cluster.local:7687", expected: "neo4j.cloud-inventory.svc.cluster.local:7687"},
		{url: "neo4j://example.com", expected: "example.com:7687"},
		{url: "neo4j+s://example.com:7687", wantErr: true},
		{url: "bolt+ssc://example.com", wantErr: true},
		{url: "://missing-host", wantErr: true},
	}
	for _, tc := range cases {
		got, err := boltSocksTarget(tc.url)
		if tc.wantErr {
			if err == nil {
				t.Errorf("expected error for %q, got %q", tc.url, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("%q: %v", tc.url, err)
			continue
		}
		if got != tc.expected {
			t.Errorf("%q: expected %q, got %q", tc.url, tc.expected, got)
		}
	}
}

type directDialer struct{}

func (directDialer) Dial(network, address string) (net.Conn, error) {
	return net.Dial(network, address)
}

// DialContext makes directDialer satisfy proxy.ContextDialer, mirroring the
// real secure-socks dialer that bridge() type-asserts to.
func (directDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	var d net.Dialer
	return d.DialContext(ctx, network, address)
}

func startEchoServer(t *testing.T) net.Addr {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_, _ = io.Copy(c, c)
			}(conn)
		}
	}()
	t.Cleanup(func() { ln.Close() })
	return ln.Addr()
}

func TestSocksRelayForwardsBothWays(t *testing.T) {
	echo := startEchoServer(t)
	relay, err := newSocksRelay(directDialer{}, echo.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(relay.close)

	conn, err := net.Dial("tcp", relay.addr())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatal(err)
	}
	if string(buf) != "ping" {
		t.Fatalf("expected %q, got %q", "ping", buf)
	}
}

func TestSocksRelayCloseStopsListener(t *testing.T) {
	echo := startEchoServer(t)
	relay, err := newSocksRelay(directDialer{}, echo.String())
	if err != nil {
		t.Fatal(err)
	}
	relay.close()
	if _, err := net.Dial("tcp", relay.addr()); err == nil {
		t.Fatal("expected dial to closed relay to fail")
	}
}

func TestSettingsParseEnableSecureSocksProxy(t *testing.T) {
	s := backend.DataSourceInstanceSettings{}
	s.JSONData = []byte(`{"url":"neo4j://localhost:7687","enableSecureSocksProxy":true}`)
	parsed, err := models.LoadPluginSettings(s)
	if err != nil {
		t.Fatal(err)
	}
	if !parsed.EnableSecureSocksProxy {
		t.Error("expected EnableSecureSocksProxy to be true")
	}
}

func TestCreateDriverDirectWhenProxyDisabled(t *testing.T) {
	s := &models.PluginSettings{URL: "neo4j://localhost:7687", Username: "neo4j"}
	s.Secrets = &models.SecretPluginSettings{Password: "x"}
	driver, relay, err := createDriver(context.Background(), s, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = driver.Close(context.Background()) })
	if relay != nil {
		t.Error("expected direct connection, got a SOCKS relay")
	}
}

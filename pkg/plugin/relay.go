package plugin

import (
	"context"
	"errors"
	"io"
	"net"
	"net/url"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"golang.org/x/net/proxy"
)

// boltSocksTarget validates a Neo4j URL for SOCKS relay mode and returns the
// host:port the relay dials through the SOCKS proxy (default port 7687).
// Driver-level TLS schemes (bolt+s, neo4j+s, ...) are rejected: the SOCKS
// tunnel is already encrypted end-to-end.
func boltSocksTarget(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "bolt+s", "bolt+ssc", "neo4j+s", "neo4j+ssc":
		return "", errors.New("driver-TLS Bolt URLs (e.g. neo4j+s://) are not supported with the secure SOCKS proxy: the proxy tunnel is already encrypted")
	}
	host := u.Hostname()
	if host == "" {
		return "", errors.New("no host in url: " + rawURL)
	}
	if u.Port() == "" {
		return net.JoinHostPort(host, "7687"), nil
	}
	return net.JoinHostPort(host, u.Port()), nil
}

// socksRelay bridges local TCP connections to a remote address through a
// SOCKS dialer. The Neo4j Go driver offers no custom-dialer hook, so when
// Grafana's secure SOCKS proxy (PDC) is enabled the driver connects to this
// relay instead, and the relay forwards every connection through the SOCKS
// tunnel to the real database address.
type socksRelay struct {
	listener net.Listener
	dialer   proxy.Dialer
	target   string
	ctx      context.Context
	cancel   context.CancelFunc
}

func newSocksRelay(dialer proxy.Dialer, target string) (*socksRelay, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	r := &socksRelay{listener: listener, dialer: dialer, target: target, ctx: ctx, cancel: cancel}
	go r.acceptLoop()
	return r, nil
}

// addr returns the loopback host:port the driver should connect to.
func (r *socksRelay) addr() string { return r.listener.Addr().String() }

func (r *socksRelay) close() {
	r.cancel()
	_ = r.listener.Close()
}

func (r *socksRelay) acceptLoop() {
	for {
		conn, err := r.listener.Accept()
		if err != nil {
			return // listener closed
		}
		go r.bridge(conn)
	}
}

func (r *socksRelay) bridge(client net.Conn) {
	defer client.Close()
	contextDialer, ok := r.dialer.(proxy.ContextDialer)
	if !ok {
		log.DefaultLogger.Error("SOCKS dialer does not support contexts")
		return
	}
	remote, err := contextDialer.DialContext(r.ctx, "tcp", r.target)
	if err != nil {
		log.DefaultLogger.Error("SOCKS relay connection failed", "target", r.target, "err", err.Error())
		return
	}
	defer remote.Close()
	done := make(chan struct{}, 2)
	go func() {
		_, _ = io.Copy(remote, client)
		done <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(client, remote)
		done <- struct{}{}
	}()
	<-done
}

package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/netbirdio/netbird/management/internals/server"
	"github.com/netbirdio/netbird/native/conduit/control"
)

const (
	conduitStatusEnabledEnv = "GOREECLOUD_CONDUIT_STATUS_ENABLED"
	conduitStatusAddrEnv    = "GOREECLOUD_CONDUIT_STATUS_ADDR"
	conduitStatusDefaultAddr = "127.0.0.1:9097"
	conduitStatusPath        = "/goreecloud/conduit/v1/status"
)

type conduitStatusSettings struct {
	enabled bool
	addr    string
}

func conduitStatusSettingsFromEnv() conduitStatusSettings {
	addr := strings.TrimSpace(os.Getenv(conduitStatusAddrEnv))
	if addr == "" {
		addr = conduitStatusDefaultAddr
	}
	return conduitStatusSettings{
		enabled: strings.TrimSpace(os.Getenv(conduitStatusEnabledEnv)) == "true",
		addr:    addr,
	}
}

func validateConduitStatusAddr(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("conduit status: invalid listen address: %w", err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return errors.New("conduit status: listener must use an explicit loopback IP address")
	}
	return nil
}

type inheritedRuntimeProvider struct{}

func (inheritedRuntimeProvider) Status(context.Context) (control.Status, error) {
	return control.Status{
		Schema:                      control.SchemaV1,
		GeneratedAt:                 time.Now().UTC(),
		Authority:                   control.AuthorityInherited,
		MigrationStage:              "implementation",
		CompatibilityBridgeActive:   true,
		ProductionCutoverAuthorized: false,
	}, nil
}

type conduitStatusServer struct {
	inner    server.Server
	settings conduitStatusSettings

	mu       sync.Mutex
	http     *http.Server
	listener net.Listener
	errors   chan error
}

func newConduitStatusServer(inner server.Server, settings conduitStatusSettings) server.Server {
	if !settings.enabled {
		return inner
	}
	return &conduitStatusServer{
		inner:    inner,
		settings: settings,
		errors:   make(chan error, 4),
	}
}

func (s *conduitStatusServer) Start(ctx context.Context) error {
	if err := validateConduitStatusAddr(s.settings.addr); err != nil {
		return err
	}
	if err := s.inner.Start(ctx); err != nil {
		return err
	}

	listener, err := net.Listen("tcp", s.settings.addr)
	if err != nil {
		_ = s.inner.Stop()
		return fmt.Errorf("conduit status: listen: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle(conduitStatusPath, control.StatusHandler{
		Provider: control.CompatibilityBridge{Inherited: inheritedRuntimeProvider{}},
	})

	httpServer := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	s.mu.Lock()
	s.listener = listener
	s.http = httpServer
	s.mu.Unlock()

	go s.forwardInnerErrors(ctx)
	go func() {
		if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) && ctx.Err() == nil {
			s.reportError(err)
		}
	}()
	return nil
}

func (s *conduitStatusServer) forwardInnerErrors(ctx context.Context) {
	select {
	case err := <-s.inner.Errors():
		if err != nil {
			s.reportError(err)
		}
	case <-ctx.Done():
	}
}

func (s *conduitStatusServer) reportError(err error) {
	select {
	case s.errors <- err:
	default:
	}
}

func (s *conduitStatusServer) Stop() error {
	s.mu.Lock()
	httpServer := s.http
	s.mu.Unlock()
	if httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = httpServer.Shutdown(ctx)
		cancel()
	}
	return s.inner.Stop()
}

func (s *conduitStatusServer) Errors() <-chan error {
	return s.errors
}

func (s *conduitStatusServer) GetContainer(key string) (any, bool) {
	return s.inner.GetContainer(key)
}

func (s *conduitStatusServer) SetContainer(key string, value any) {
	s.inner.SetContainer(key, value)
}

func init() {
	baseNewServer := newServer
	newServer = func(cfg *server.Config) server.Server {
		return newConduitStatusServer(baseNewServer(cfg), conduitStatusSettingsFromEnv())
	}
}

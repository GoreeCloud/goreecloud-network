package cmd

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	goreecloudstatus "github.com/netbirdio/netbird/goreecloud/status"
	"github.com/netbirdio/netbird/management/internals/server"
)

const (
	goreecloudNetworkStatusFileEnv  = "GOREECLOUD_NETWORK_STATUS_FILE"
	goreecloudNetworkStatusInterval = 30 * time.Second
)

var goreecloudStatusAdapterOnce sync.Once

// EnableGoreeCloudStatusAdapter wraps management server construction with the
// fork-only, privacy-minimized local status adapter.  It is safe to call more
// than once and adds no listener or API dependency.
func EnableGoreeCloudStatusAdapter() {
	goreecloudStatusAdapterOnce.Do(func() {
		baseNewServer := newServer
		newServer = func(cfg *server.Config) server.Server {
			return newGoreeCloudStatusServer(baseNewServer(cfg), os.Getenv(goreecloudNetworkStatusFileEnv))
		}
	})
}

type goreecloudStatusServer struct {
	inner server.Server
	path  string

	running atomic.Bool
	cancel  context.CancelFunc
}

func newGoreeCloudStatusServer(inner server.Server, path string) server.Server {
	return &goreecloudStatusServer{inner: inner, path: path}
}

func (s *goreecloudStatusServer) Start(ctx context.Context) error {
	if s.path != "" {
		s.publish(time.Now())
	}

	if err := s.inner.Start(ctx); err != nil {
		s.running.Store(false)
		if s.path != "" {
			s.publish(time.Now())
		}

		return err
	}

	s.running.Store(true)
	if s.path != "" {
		s.publish(time.Now())
		publisherCtx, cancel := context.WithCancel(ctx)
		s.cancel = cancel
		go s.publishLoop(publisherCtx)
	}

	return nil
}

func (s *goreecloudStatusServer) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	s.running.Store(false)

	err := s.inner.Stop()
	if s.path != "" {
		s.publish(time.Now())
	}

	return err
}

func (s *goreecloudStatusServer) Errors() <-chan error {
	return s.inner.Errors()
}

func (s *goreecloudStatusServer) GetContainer(key string) (any, bool) {
	return s.inner.GetContainer(key)
}

func (s *goreecloudStatusServer) SetContainer(key string, container any) {
	s.inner.SetContainer(key, container)
}

func (s *goreecloudStatusServer) publishLoop(ctx context.Context) {
	ticker := time.NewTicker(goreecloudNetworkStatusInterval)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			s.publish(now)
		case <-ctx.Done():
			return
		}
	}
}

func (s *goreecloudStatusServer) publish(now time.Time) {
	evidence := goreecloudstatus.RuntimeEvidence{}
	if s.running.Load() {
		// A successful management Server.Start means the management store,
		// managers, gRPC/API server, and listening socket were initialized.
		// That is sufficient to prove the coordination and policy-service
		// boundaries, but not end-to-end private connectivity or Network DNS.
		evidence.PeerCoordinationReady = true
		evidence.AccessPolicyReady = true
	}

	if err := goreecloudstatus.WriteFile(s.path, goreecloudstatus.SnapshotFromEvidence(now, evidence)); err != nil {
		log.Warnf("goreecloud status handoff failed: %v", err)
	}
}

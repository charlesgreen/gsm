// Package gsmtest provides an in-process Google Secret Manager emulator for Go
// tests, as a parallel to the bttest and pstest harnesses Google ships. It runs
// the emulator over TCP, an in-memory pipe, or persistent storage, and hands back
// a Secret Manager client wired to it.
package gsmtest

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/akutz/memconn"
	"github.com/charlesgreen/gsm/internal/api/routes"
	"github.com/charlesgreen/gsm/internal/storage"
	"google.golang.org/api/option"
)

// Option configures the emulator created by New.
type Option func(*options)

// StorageFile enables persistent secret storage using a file path
//
// Previous values are consumed from the file if they exist.
func StorageFile(path string) Option {
	return func(o *options) {
		o.storageFile = path
	}
}

// Listener overrides where requests are served from.
func Listener(lis net.Listener) Option {
	return func(o *options) {
		o.listener = lis
	}
}

// Addr the server should listen on
//
// Defaults to a random port of localhost
func Addr(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

// InMemory configures the server to use a local buffer for transport instead of
// network sockets. This allows it to be used with [testing/synctest] or generally
// faster startup times.
func InMemory() Option {
	return func(o *options) {
		o.inMemory = true
	}
}

// ShutdownTimeout when waiting for in-flight requests to finish
//
// Defaults to 1s
func ShutdownTimeout(dur time.Duration) Option {
	return func(o *options) {
		o.shutdownTimeout = dur
	}
}

// New instance for emulating the Google Secret Manager
func New(t testing.TB, opts ...Option) (*SecretManager, error) {
	var options options
	for _, o := range opts {
		o(&options)
	}

	lis, err := options.createListener()
	if err != nil {
		return nil, fmt.Errorf("creating listener: %w", err)
	}

	store, err := options.createStore(t)
	if err != nil {
		return nil, fmt.Errorf("creating store: %w", err)
	}

	srv := &http.Server{Handler: routes.SetupRoutes(store)}
	return &SecretManager{
		tb:              t,
		srv:             srv,
		lis:             lis,
		store:           store,
		shutdownTimeout: cmp.Or(options.shutdownTimeout, time.Second),
	}, nil
}

// SecretManager is a running emulator instance and the handle tests use to start
// it and obtain a client.
type SecretManager struct {
	tb              testing.TB
	srv             *http.Server
	lis             net.Listener
	store           storage.Storage
	shutdownTimeout time.Duration
}

// Start the server and block until finished. The context is used for cancellation.
func (s *SecretManager) Start(ctx context.Context) error {
	// Shut the server down when the context is canceled. The channel is blocked until
	// all in-flight requests have finished processing.
	done := make(chan struct{})
	go func() {
		<-ctx.Done()

		// Unblock exiting the parent func when requests have drained
		defer close(done)

		// Perhaps overkill, but avoid hanging someone's test indefinitely.
		timeCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		if err := s.srv.Shutdown(timeCtx); err != nil {
			s.tb.Logf("error shutting down: %v", err)
		}
	}()

	// Block until it shuts down from context or otherwise
	err := s.srv.Serve(s.lis)

	// Wait until graceful shutdown is complete before exiting
	if errors.Is(err, http.ErrServerClosed) {
		<-done
		return nil
	}
	return err
}

// Addr the server is listening on
func (s *SecretManager) Addr() string {
	return s.lis.Addr().String()
}

// Endpoint the client options expect
func (s *SecretManager) Endpoint() string {
	return "http://" + s.Addr()
}

// Client connected to the local emulator
func (s *SecretManager) Client(ctx context.Context) (*secretmanager.Client, error) {
	if _, ok := s.lis.(*memconn.Listener); ok {
		return s.memClient(ctx)
	}
	return secretmanager.NewRESTClient(
		ctx,
		option.WithoutAuthentication(),
		option.WithEndpoint(s.Endpoint()),
	)
}

func (s *SecretManager) memClient(ctx context.Context) (*secretmanager.Client, error) {
	dial := func(ctx context.Context, _, _ string) (net.Conn, error) {
		return memconn.DialContext(ctx, "memu", s.Addr())
	}
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: dial,
			// All other HTTP options are ignored when providing a custom client, so we
			// need to ignore https ourselves.
			DialTLSContext: dial,
		},
	}
	return secretmanager.NewRESTClient(ctx, option.WithHTTPClient(client))
}

type options struct {
	addr            string
	inMemory        bool
	listener        net.Listener
	storageFile     string
	shutdownTimeout time.Duration
}

func (o options) createListener() (net.Listener, error) {
	if o.listener != nil {
		return o.listener, nil
	}
	if o.inMemory {
		return memconn.Listen("memu", strconv.FormatInt(rand.Int63(), 16))
	}
	if o.addr != "" {
		return net.Listen("tcp", o.addr)
	}
	return net.Listen("tcp", "localhost:0")
}

func (o options) createStore(t testing.TB) (storage.Storage, error) {
	if o.storageFile != "" {
		return persistentStore(t, o.storageFile)
	}

	return storage.NewMemoryStorage(), nil
}

func persistentStore(t testing.TB, path string) (*storage.PersistentStorage, error) {
	store, err := storage.NewPersistentStorage(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create persistent storage: %w", err)
	}
	if err := store.Load(); err != nil {
		t.Logf("warning: failed loading existing storage: %v", err)
	}
	return store, nil
}

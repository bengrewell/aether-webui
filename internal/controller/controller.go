package controller

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/provider/meta"
	"github.com/bengrewell/aether-webui/internal/security"
	"github.com/bengrewell/aether-webui/internal/store"
)

// ProviderFactory creates a provider from a store client and pre-wired options.
// Factories are registered via WithProvider and called during Run().
type ProviderFactory func(ctx context.Context, st store.Client, opts []provider.Option) (provider.Provider, error)

// providerReg holds a pending provider registration.
type providerReg struct {
	name    string
	enabled bool
	factory ProviderFactory
}

// Controller orchestrates server lifecycle: logging, TLS, store, transports,
// providers, frontend, HTTP server, and graceful shutdown.
type Controller struct {
	// Configuration (set by options, immutable after New)
	listenAddr       string
	debug            bool
	dataDir          string
	versionInfo      meta.VersionInfo
	tlsAuto          bool
	tlsCert          string
	tlsKey           string
	tlsMTLSCA        string
	apiToken         string
	rbacEnabled      bool
	frontendEnabled  bool
	frontendDir      string
	metricsInterval  string
	metricsRetention string
	encryptionKey    string
	storeOpts        []store.Option
	providerRegs     []providerReg

	// Runtime state (populated during Run)
	log        *slog.Logger
	store      store.Client
	transports []Transport
	providers  []provider.Provider
	server     *http.Server
	tlsResult  *security.TLSResult
}

// New creates a Controller configured by the given options.
// Defaults are applied for any unset fields.
func New(opts ...Option) (*Controller, error) {
	c := &Controller{
		listenAddr:       "127.0.0.1:8186",
		dataDir:          "/var/lib/aether-webd",
		metricsInterval:  "10s",
		metricsRetention: "24h",
		frontendEnabled:  true,
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

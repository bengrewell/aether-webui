package rest

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/store"
)

// Config holds the parameters needed to construct a REST transport.
type Config struct {
	APITitle   string
	APIVersion string
	Log        *slog.Logger // scoped logger for transport-level events
	Store      store.Client // shared store for providers
}

// Transport owns the Chi router, Huma API, and shared dependencies that
// providers need (logger, store). It replaces the inline setup previously
// done in main.go.
type Transport struct {
	router chi.Router
	api    huma.API
	log    *slog.Logger
	store  store.Client
}

// NewTransport creates a Transport with the given config and optional Chi
// middleware applied in order.
func NewTransport(cfg Config, middleware ...func(http.Handler) http.Handler) *Transport {
	r := chi.NewMux()
	for _, mw := range middleware {
		r.Use(mw)
	}
	log := cfg.Log
	if log == nil {
		log = slog.Default()
	}
	api := humachi.New(r, huma.DefaultConfig(cfg.APITitle, cfg.APIVersion))
	return &Transport{router: r, api: api, log: log, store: cfg.Store}
}

func (t *Transport) API() huma.API         { return t.api }
func (t *Transport) Handler() http.Handler { return t.router }
func (t *Transport) Log() *slog.Logger     { return t.log }
func (t *Transport) Store() store.Client   { return t.store }

// Mount attaches a catch-all handler (e.g. frontend SPA) after API routes.
func (t *Transport) Mount(pattern string, h http.Handler) {
	t.router.Handle(pattern, h)
}

// ProviderOpts returns the common provider.Option set (Huma API, logger, store)
// so each provider constructor doesn't have to wire them individually.
func (t *Transport) ProviderOpts(providerName string) []provider.Option {
	return []provider.Option{
		provider.WithHuma(t.api),
		provider.WithLogger(t.log.With("provider", providerName)),
		provider.WithStore(t.store),
	}
}

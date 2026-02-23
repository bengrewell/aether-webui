package rest

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/store"
)

const apiOverview = `Backend API service for the **Aether WebUI**. This service manages Aether 5G deployments including SD-Core components, gNBs (srsRAN, OCUDU), Kubernetes clusters, and host systems.

## Base URL

All API endpoints are served under ` + "`/api/v1/`" + `.

## Authentication

When token authentication is enabled (` + "`--api-token`" + ` flag or ` + "`AETHER_API_TOKEN`" + ` env var), all ` + "`/api/*`" + ` endpoints require a bearer token:

` + "```" + `
Authorization: Bearer <token>
` + "```" + `

The following paths are always public: ` + "`/healthz`" + `, ` + "`/openapi.json`" + `, ` + "`/docs`" + `, and frontend static files.

## Providers

The API is organized into **providers** — modular units that each register a set of endpoints:

| Provider | Prefix | Description |
|----------|--------|-------------|
| meta | ` + "`/api/v1/meta/`" + ` | Server introspection — version, build info, runtime, config, provider status, store health |

## Response Format

All endpoints return JSON. Error responses follow the [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457) Problem Details format.
`

// Config holds the parameters needed to construct a REST transport.
type Config struct {
	APITitle         string
	APIVersion       string
	Log              *slog.Logger // scoped logger for transport-level events
	Store            store.Client // shared store for providers
	TokenAuthEnabled bool         // when true, adds Bearer security scheme to the OpenAPI spec
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
	humaConfig := huma.DefaultConfig(cfg.APITitle, cfg.APIVersion)
	humaConfig.Info.Description = apiOverview
	// Disable Huma's built-in docs handler; we register a custom one below
	// that points to the OpenAPI 3.0 spec to avoid Stoplight Elements crashes.
	humaConfig.DocsPath = ""

	if cfg.TokenAuthEnabled {
		if humaConfig.Components == nil {
			humaConfig.Components = &huma.Components{}
		}
		if humaConfig.Components.SecuritySchemes == nil {
			humaConfig.Components.SecuritySchemes = make(map[string]*huma.SecurityScheme)
		}
		humaConfig.Components.SecuritySchemes["bearerAuth"] = &huma.SecurityScheme{
			Type:         "http",
			Scheme:       "bearer",
			Description:  "Bearer token authentication. Pass the token configured via --api-token or AETHER_API_TOKEN.",
			BearerFormat: "token",
		}
		humaConfig.Security = []map[string][]string{
			{"bearerAuth": {}},
		}
	}

	api := humachi.New(r, humaConfig)

	// Serve the docs page using the downgraded OpenAPI 3.0 spec. Huma's
	// default serves 3.1 which causes Stoplight Elements v9.0.0 to crash on
	// endpoints that define path/query parameters.
	docsTitle := cfg.APITitle + " Reference"
	r.Get("/docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="referrer" content="same-origin" />
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no" />
    <title>%s</title>
    <link href="https://unpkg.com/@stoplight/elements@9.0.0/styles.min.css" rel="stylesheet" />
    <script src="https://unpkg.com/@stoplight/elements@9.0.0/web-components.min.js" integrity="sha256-Tqvw1qE2abI+G6dPQBc5zbeHqfVwGoamETU3/TSpUw4="
            crossorigin="anonymous"></script>
  </head>
  <body style="height: 100vh;">
    <elements-api
      apiDescriptionUrl="/openapi-3.0.yaml"
      router="hash"
      layout="sidebar"
      tryItCredentialsPolicy="same-origin"
    />
  </body>
</html>`, docsTitle)
	})

	return &Transport{router: r, api: api, log: log, store: cfg.Store}
}

func (t *Transport) API() huma.API         { return t.api }
func (t *Transport) Handler() http.Handler { return t.router }
func (t *Transport) Log() *slog.Logger     { return t.log }
func (t *Transport) Store() store.Client   { return t.store }

// HandleFunc registers a plain http.HandlerFunc on the router, outside of the
// Huma/OpenAPI spec. Useful for operational endpoints like /healthz.
func (t *Transport) HandleFunc(pattern string, fn http.HandlerFunc) {
	t.router.HandleFunc(pattern, fn)
}

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

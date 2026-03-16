package controller

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rs/cors"

	"github.com/bengrewell/aether-webui/internal/api/rest"
	"github.com/bengrewell/aether-webui/internal/auth"
	"github.com/bengrewell/aether-webui/internal/frontend"
	"github.com/bengrewell/aether-webui/internal/logging"
	mcpserver "github.com/bengrewell/aether-webui/internal/mcp"
	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/provider/meta"
	"github.com/bengrewell/aether-webui/internal/provider/nodes"
	"github.com/bengrewell/aether-webui/internal/provider/onramp"
	"github.com/bengrewell/aether-webui/internal/provider/system"
	"github.com/bengrewell/aether-webui/internal/security"
	"github.com/bengrewell/aether-webui/internal/store"
)

// Run executes the full server lifecycle: logging, TLS, store, transports,
// providers, frontend, HTTP server, and graceful shutdown. It blocks until
// ctx is cancelled or an OS signal triggers shutdown.
func (c *Controller) Run(ctx context.Context) error {
	c.setupLogging()

	c.log.Info("aether-webd starting",
		"version", c.versionInfo.Version,
		"build_date", c.versionInfo.BuildDate,
		"branch", c.versionInfo.Branch,
		"commit", c.versionInfo.CommitHash,
		"debug", c.debug,
	)

	if err := c.buildTLS(); err != nil {
		c.log.Error("TLS configuration failed", "error", err)
		return fmt.Errorf("tls: %w", err)
	}

	if err := c.openStore(ctx); err != nil {
		c.log.Error("store open failed", "path", c.dataDir, "err", err)
		return fmt.Errorf("store: %w", err)
	}
	defer c.store.Close()

	mw := c.buildMiddleware()

	transport := c.createRESTTransport(mw)

	c.registerHealthz(transport, time.Now())

	if err := c.initProviders(ctx, transport); err != nil {
		return fmt.Errorf("providers: %w", err)
	}

	if c.mcpEnabled {
		if err := c.startMCP(ctx); err != nil {
			return fmt.Errorf("mcp: %w", err)
		}
	}

	c.mountFrontend(transport)

	c.server = &http.Server{
		Addr:    c.listenAddr,
		Handler: transport.Handler(),
	}

	serverErr := make(chan error, 1)
	go func() {
		if c.tlsResult != nil {
			c.server.TLSConfig = c.tlsResult.Config
			c.log.Info("starting HTTPS server", "addr", c.listenAddr)
			if err := c.server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				c.log.Error("HTTPS server error", "error", err)
				serverErr <- err
			}
		} else {
			c.log.Info("starting HTTP server", "addr", c.listenAddr)
			if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				c.log.Error("HTTP server error", "error", err)
				serverErr <- err
			}
		}
	}()

	return c.awaitShutdown(ctx, serverErr)
}

// setupLogging initializes structured logging via the logging package.
// When MCP stdio transport is active (mcpEnabled && no mcpListenAddr),
// log output goes to stderr to avoid conflicting with JSON-RPC on stdout.
func (c *Controller) setupLogging() {
	level := slog.LevelInfo
	if c.debug {
		level = slog.LevelDebug
	}
	opts := logging.Options{
		Level:     level,
		AddSource: c.debug,
	}
	if c.mcpEnabled && c.mcpListenAddr == "" {
		opts.Writer = os.Stderr
	}
	logging.Setup(opts)
	c.log = slog.Default()
}

// buildTLS resolves TLS configuration from the controller's TLS fields.
func (c *Controller) buildTLS() error {
	tlsEnabled := c.tlsAuto || (c.tlsCert != "" && c.tlsKey != "") || c.tlsMTLSCA != ""
	if !tlsEnabled {
		return nil
	}
	result, err := security.BuildTLSConfig(security.TLSOptions{
		AutoTLS:    c.tlsAuto,
		DataDir:    c.dataDir,
		CertFile:   c.tlsCert,
		KeyFile:    c.tlsKey,
		MTLSCAFile: c.tlsMTLSCA,
	})
	if err != nil {
		return err
	}
	c.tlsResult = result

	logAttrs := []any{
		"cert_source", result.CertSource,
		"mtls", result.MTLSEnabled,
	}
	if result.CertDir != "" {
		logAttrs = append(logAttrs, "cert_dir", result.CertDir)
	}
	c.log.Info("TLS enabled", logAttrs...)
	return nil
}

// openStore opens the SQLite database at dataDir/app.db.
func (c *Controller) openStore(ctx context.Context) error {
	dbPath := filepath.Join(c.dataDir, "app.db")
	st, err := store.New(ctx, dbPath, c.storeOpts...)
	if err != nil {
		return err
	}
	c.store = st
	return nil
}

// buildMiddleware assembles the middleware chain (CORS + logging + optional token auth).
func (c *Controller) buildMiddleware() []func(http.Handler) http.Handler {
	var mw []func(http.Handler) http.Handler
	if len(c.corsOrigins) > 0 {
		corsMiddleware := cors.New(cors.Options{
			AllowedOrigins:   c.corsOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Authorization", "Content-Type"},
			AllowCredentials: true,
			MaxAge:           300,
		})
		mw = append(mw, corsMiddleware.Handler)
		c.log.Info("CORS enabled", "origins", c.corsOrigins)
	}
	mw = append(mw, logging.RequestLogger())
	if c.apiToken != "" {
		mw = append(mw, auth.TokenAuth(c.apiToken, auth.DefaultSkipPaths))
		c.log.Info("token authentication enabled")
	}
	return mw
}

// createRESTTransport creates the REST transport and adds it to c.transports.
func (c *Controller) createRESTTransport(mw []func(http.Handler) http.Handler) *rest.Transport {
	transport := rest.NewTransport(rest.Config{
		APITitle:         "Aether WebUI API",
		APIVersion:       c.versionInfo.Version,
		Log:              c.log,
		Store:            c.store,
		TokenAuthEnabled: c.apiToken != "",
	}, mw...)
	c.transports = append(c.transports, transport)
	return transport
}

// initProviders initializes registered provider factories and the meta provider.
func (c *Controller) initProviders(ctx context.Context, transport *rest.Transport) error {
	for _, reg := range c.providerRegs {
		var allOpts []provider.Option
		for _, t := range c.transports {
			allOpts = append(allOpts, t.ProviderOpts(reg.name)...)
		}
		p, err := reg.factory(ctx, c.store, allOpts)
		if err != nil {
			return fmt.Errorf("provider %q: %w", reg.name, err)
		}
		if !reg.enabled {
			p.Disable()
		}
		c.providers = append(c.providers, p)
	}

	metaProvider := c.createMetaProvider(transport)
	c.providers = append(c.providers, metaProvider)

	// Start enabled providers.
	for _, p := range c.providers {
		si := p.(interface{ StatusInfo() provider.StatusInfo }).StatusInfo()
		if si.Enabled {
			if err := p.Start(); err != nil {
				return fmt.Errorf("provider %q start: %w", p.Name(), err)
			}
		}
	}
	return nil
}

// createMetaProvider builds the meta provider with closures that capture runtime state.
func (c *Controller) createMetaProvider(transport *rest.Transport) *meta.Meta {
	frontendSource := ""
	if c.frontendEnabled {
		if c.frontendDir != "" {
			frontendSource = "directory"
		} else {
			frontendSource = "embedded"
		}
	}

	appConfig := meta.AppConfig{
		ListenAddress: c.listenAddr,
		DebugEnabled:  c.debug,
		Security: meta.SecurityConfig{
			TLSEnabled:       c.tlsResult != nil,
			TLSAutoGenerated: c.tlsResult != nil && c.tlsResult.AutoCert,
			MTLSEnabled:      c.tlsResult != nil && c.tlsResult.MTLSEnabled,
			TokenAuthEnabled: c.apiToken != "",
			RBACEnabled:      c.rbacEnabled,
			CORSOrigins:      c.corsOrigins,
		},
		Frontend: meta.FrontendConfig{
			Enabled: c.frontendEnabled,
			Source:  frontendSource,
			Dir:     c.frontendDir,
		},
		Storage: meta.StorageConfig{
			DataDir: c.dataDir,
		},
		Metrics: meta.MetricsConfig{
			Interval:  c.metricsInterval,
			Retention: c.metricsRetention,
		},
	}

	schemaVerFn := func() (int, error) {
		return c.store.GetSchemaVersion()
	}

	// Closure captures c.providers by reference — safe since providersFn is
	// only called at request time, after all providers are fully constructed.
	providersFn := func() []meta.ProviderStatus {
		type statusInfoer interface {
			StatusInfo() provider.StatusInfo
		}
		out := make([]meta.ProviderStatus, len(c.providers))
		for i, p := range c.providers {
			si := p.(statusInfoer).StatusInfo()
			out[i] = meta.ProviderStatus{
				Name:           p.Name(),
				Enabled:        si.Enabled,
				Running:        si.Running,
				EndpointCount:  si.EndpointCount,
				Degraded:       si.Degraded,
				DegradedReason: si.DegradedReason,
			}
		}
		return out
	}

	storeInfoFn := c.buildStoreInfoFn()

	return meta.NewProvider(
		c.versionInfo,
		appConfig,
		schemaVerFn,
		providersFn,
		storeInfoFn,
		transport.ProviderOpts("meta")...,
	)
}

// buildStoreInfoFn returns a closure that produces live store diagnostics.
func (c *Controller) buildStoreInfoFn() meta.StoreInfoFunc {
	return func(ctx context.Context) meta.StoreInfo {
		info := meta.StoreInfo{
			Engine: "sqlite",
			Path:   c.store.Path(),
		}
		if fi, err := os.Stat(c.store.Path()); err == nil {
			info.FileSizeBytes = fi.Size()
		}
		if v, err := c.store.GetSchemaVersion(); err == nil {
			info.SchemaVersion = v
		}

		diagKey := store.Key{Namespace: "_diagnostics", ID: "healthcheck"}
		checks := []meta.DiagnosticCheck{
			runCheck("ping", func() error { return c.store.Health(ctx) }),
			runCheck("write", func() error {
				_, err := store.Save(c.store, ctx, diagKey, "ok")
				return err
			}),
			runCheck("read", func() error {
				item, ok, err := store.Load[string](c.store, ctx, diagKey)
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("key not found")
				}
				if item.Data != "ok" {
					return fmt.Errorf("data mismatch: got %q, want %q", item.Data, "ok")
				}
				return nil
			}),
			runCheck("delete", func() error { return c.store.Delete(ctx, diagKey) }),
		}
		info.Diagnostics = checks

		passed := 0
		for _, ch := range checks {
			if ch.Passed {
				passed++
			}
		}
		switch {
		case passed == len(checks):
			info.Status = "healthy"
		case passed > 0:
			info.Status = "degraded"
		default:
			info.Status = "unhealthy"
		}
		return info
	}
}

// startMCP creates the MCP server and starts the configured transports.
func (c *Controller) startMCP(ctx context.Context) error {
	// Find the concrete provider instances from c.providers.
	var (
		nodesProvider  *nodes.Nodes
		onrampProvider *onramp.OnRamp
		systemProvider *system.System
		metaProvider   *meta.Meta
	)
	for _, p := range c.providers {
		switch v := p.(type) {
		case *nodes.Nodes:
			nodesProvider = v
		case *onramp.OnRamp:
			onrampProvider = v
		case *system.System:
			systemProvider = v
		case *meta.Meta:
			metaProvider = v
		}
	}

	srv := mcpserver.New(mcpserver.Config{
		Store:   c.store,
		Nodes:   nodesProvider,
		OnRamp:  onrampProvider,
		System:  systemProvider,
		Meta:    metaProvider,
		Log:     c.log.With("component", "mcp"),
		Version: c.versionInfo.Version,
	})

	// Stdio transport: run in a goroutine, blocks until ctx done.
	if c.mcpListenAddr == "" {
		go func() {
			if err := srv.RunStdio(ctx); err != nil {
				c.log.Error("MCP stdio server error", "error", err)
			}
		}()
		c.log.Info("MCP server started on stdio")
	}

	// StreamableHTTP transport on a separate address.
	if c.mcpListenAddr != "" {
		c.mcpServer = &http.Server{
			Addr:    c.mcpListenAddr,
			Handler: srv.HTTPHandler(),
		}
		go func() {
			c.log.Info("starting MCP HTTP server", "addr", c.mcpListenAddr)
			if err := c.mcpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				c.log.Error("MCP HTTP server error", "error", err)
			}
		}()
	}

	return nil
}

// mountFrontend attaches the SPA handler to the first transport, if enabled.
func (c *Controller) mountFrontend(transport Transport) {
	if !c.frontendEnabled {
		return
	}
	var h http.Handler
	if c.frontendDir != "" {
		c.log.Info("serving frontend from directory", "path", c.frontendDir)
		h = frontend.NewHandler(os.DirFS(c.frontendDir), "")
	} else {
		c.log.Info("serving frontend from embedded files")
		h = frontend.NewHandler(frontend.DistFS, "dist")
	}
	transport.Mount("/*", h)
}

// awaitShutdown blocks until ctx is cancelled or an OS signal triggers shutdown.
// SIGTERM always triggers immediate shutdown (systemd sends this on stop).
// SIGINT uses a double-press confirmation: first press warns, second shuts down.
func (c *Controller) awaitShutdown(ctx context.Context, serverErr <-chan error) error {
	intChan := make(chan os.Signal, 1)
	termChan := make(chan os.Signal, 1)
	signal.Notify(intChan, syscall.SIGINT)
	signal.Notify(termChan, syscall.SIGTERM)
	defer signal.Stop(intChan)
	defer signal.Stop(termChan)

	var lastInterrupt time.Time
	confirmWindow := 3 * time.Second

	for {
		select {
		case err := <-serverErr:
			return err
		case <-ctx.Done():
			return c.shutdown()
		case <-termChan:
			c.log.Info("SIGTERM received, shutting down")
			return c.shutdown()
		case <-intChan:
			now := time.Now()
			if now.Sub(lastInterrupt) <= confirmWindow {
				return c.shutdown()
			}
			lastInterrupt = now
			c.log.Warn("interrupt received, press Ctrl+C again within 3 seconds to quit")
		}
	}
}

// shutdown performs graceful server shutdown and stops all providers in reverse order.
func (c *Controller) shutdown() error {
	c.log.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var firstErr error
	if err := c.server.Shutdown(shutdownCtx); err != nil {
		c.log.Error("server shutdown error", "error", err)
		firstErr = err
	}

	if c.mcpServer != nil {
		if err := c.mcpServer.Shutdown(shutdownCtx); err != nil {
			c.log.Error("MCP server shutdown error", "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	// Stop providers in reverse registration order.
	for i := len(c.providers) - 1; i >= 0; i-- {
		if err := c.providers[i].Stop(); err != nil {
			c.log.Error("provider stop error", "provider", c.providers[i].Name(), "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

// runCheck executes a named diagnostic function and captures its result and latency.
func runCheck(name string, fn func() error) meta.DiagnosticCheck {
	start := time.Now()
	err := fn()
	d := time.Since(start)
	ch := meta.DiagnosticCheck{
		Name:    name,
		Passed:  err == nil,
		Latency: d.String(),
	}
	if err != nil {
		ch.Error = err.Error()
	}
	return ch
}

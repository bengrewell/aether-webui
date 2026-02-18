package main

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

	"github.com/bengrewell/aether-webui/internal/api/rest"
	"github.com/bengrewell/aether-webui/internal/auth"
	"github.com/bengrewell/aether-webui/internal/frontend"
	"github.com/bengrewell/aether-webui/internal/logging"
	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/provider/meta"
	"github.com/bengrewell/aether-webui/internal/security"
	"github.com/bengrewell/aether-webui/internal/store"
	"github.com/bgrewell/usage"
)

var (
	version    string = "dev"
	buildDate  string = "unknown"
	branch     string = "unknown"
	commitHash string = "unknown"
)

func main() {

	// Setup usage and command-line options
	u := usage.NewUsage(
		usage.WithApplicationName("aether-webd"),
		usage.WithApplicationVersion(version),
		usage.WithApplicationBuildDate(buildDate),
		usage.WithApplicationCommitHash(commitHash),
		usage.WithApplicationBranch(branch),
		usage.WithApplicationDescription("Backend API service for the Aether WebUI. This service is responsible for executing deployment tasks, gathering system information, and monitoring the health and metrics of Aether 5G deployments. It manages SD-Core components, gNBs (such as srsRAN and OCUDU), Kubernetes clusters, and host systems."),
	)

	flagVersion := u.AddBooleanOption("v", "version", false, "Print version information and exit", "", nil)
	flagDebug := u.AddBooleanOption("d", "debug", false, "Enable debug mode for verbose logging and diagnostic output", "", nil)
	flagListen := u.AddStringOption("l", "listen", "127.0.0.1:8186", "Address and port the API server will listen on (e.g., 0.0.0.0:8186 for all interfaces)", "", nil)

	secOptions := u.AddGroup(2, "Security Options", "Options that control security settings")
	flagTLS := u.AddBooleanOption("", "tls", false, "Enable TLS. Auto-generates a self-signed certificate if --tls-cert and --tls-key are not provided.", "", secOptions)
	flagTLSCert := u.AddStringOption("t", "tls-cert", "", "Path to the TLS certificate file for HTTPS. When provided with --tls-key, the server will use HTTPS instead of HTTP.", "", secOptions)
	flagTLSKey := u.AddStringOption("k", "tls-key", "", "Path to the TLS private key file for HTTPS. Required when --tls-cert is specified.", "", secOptions)
	flagMTLSCACert := u.AddStringOption("m", "mtls-ca-cert", "", "Path to the CA certificate file for verifying client certificates. Enables mutual TLS (mTLS) authentication when specified.", "", secOptions)
	flagAPIToken := u.AddStringOption("", "api-token", "", "Bearer token for API authentication. Falls back to AETHER_API_TOKEN env var. When set, all /api/* requests require Authorization: Bearer <token>.", "", secOptions)
	flagEnableRBAC := u.AddBooleanOption("r", "enable-rbac", false, "Enable role-based access control (RBAC) authentication and authorization middleware.", "", secOptions)

	exeOptions := u.AddGroup(1, "Execution Options", "Options that control API command execution")
	flagExecUser := u.AddStringOption("u", "exec-user", "", "User account under which API commands will be executed", "", exeOptions)
	flagExecEnv := u.AddStringOption("e", "exec-env", "", "Environment variables to pass to the command execution context", "", exeOptions)

	frontendOptions := u.AddGroup(3, "Frontend Options", "Options that control frontend serving")
	flagServeFrontend := u.AddBooleanOption("f", "serve-frontend", true, "Enable serving frontend static files from embedded or custom directory", "", frontendOptions)
	flagFrontendDir := u.AddStringOption("", "frontend-dir", "", "Override embedded frontend with files from this directory (for development)", "", frontendOptions)

	storageOptions := u.AddGroup(4, "Storage Options", "Options that control persistent state storage")
	flagDataDir := u.AddStringOption("", "data-dir", "/var/lib/aether-webd", "Directory for persistent state database", "", storageOptions)
	flagEncryptionKey := u.AddStringOption("", "encryption-key", "", "32-byte encryption key for node passwords. Falls back to AETHER_ENCRYPTION_KEY env var. Auto-generated if neither is provided.", "", secOptions)

	metricsOptions := u.AddGroup(5, "Metrics Options", "Options that control metrics collection")
	flagMetricsInterval := u.AddStringOption("", "metrics-interval", "10s", "How often to collect system metrics (e.g., '10s', '30s', '1m')", "", metricsOptions)
	flagMetricsRetention := u.AddStringOption("", "metrics-retention", "24h", "How long to retain historical metrics data (e.g., '24h', '7d')", "", metricsOptions)

	parsed := u.Parse()

	_ = flagEncryptionKey

	if !parsed {
		u.PrintUsage()
		return
	}

	if *flagVersion {
		fmt.Printf("aether-webd %s (branch: %s, commit: %s, built: %s)\n",
			version, branch, commitHash, buildDate)
		return
	}

	// Initialize structured logging
	logLevel := slog.LevelInfo
	if *flagDebug {
		logLevel = slog.LevelDebug
	}
	logging.Setup(logging.Options{
		Level:     logLevel,
		AddSource: *flagDebug,
	})
	log := slog.Default()

	log.Info("aether-webd starting",
		"version", version,
		"build_date", buildDate,
		"branch", branch,
		"commit", commitHash,
		"debug", *flagDebug,
	)

	log.Debug("unimplemented execution options",
		"exec_user", *flagExecUser,
		"exec_env", *flagExecEnv,
	)

	// Resolve API token: flag takes precedence, then env var.
	apiToken := *flagAPIToken
	if apiToken == "" {
		apiToken = os.Getenv("AETHER_API_TOKEN")
	}

	// TLS setup: --tls flag, cert+key presence, or mTLS CA all enable TLS.
	tlsEnabled := *flagTLS || (*flagTLSCert != "" && *flagTLSKey != "") || *flagMTLSCACert != ""

	var tlsResult *security.TLSResult
	if tlsEnabled {
		var err error
		tlsResult, err = security.BuildTLSConfig(security.TLSOptions{
			AutoTLS:    *flagTLS,
			DataDir:    *flagDataDir,
			CertFile:   *flagTLSCert,
			KeyFile:    *flagTLSKey,
			MTLSCAFile: *flagMTLSCACert,
		})
		if err != nil {
			log.Error("TLS configuration failed", "error", err)
			os.Exit(1)
		}
		logAttrs := []any{
			"cert_source", tlsResult.CertSource,
			"mtls", tlsResult.MTLSEnabled,
		}
		if tlsResult.CertDir != "" {
			logAttrs = append(logAttrs, "cert_dir", tlsResult.CertDir)
		}
		log.Info("TLS enabled", logAttrs...)
	}

	// Assemble middleware chain.
	var middlewares []func(http.Handler) http.Handler
	middlewares = append(middlewares, logging.RequestLogger())
	if apiToken != "" {
		middlewares = append(middlewares, auth.TokenAuth(apiToken, auth.DefaultSkipPaths))
		log.Info("token authentication enabled")
	}

	// Open database using --data-dir flag
	dbPath := filepath.Join(*flagDataDir, "app.db")
	dbcli, err := store.New(context.Background(), dbPath)
	if err != nil {
		log.Error("store open failed", "path", *flagDataDir, "err", err)
		os.Exit(1)
	}

	// Create REST transport (Chi router + Huma API + shared deps)
	transport := rest.NewTransport(rest.Config{
		APITitle:         "Aether WebUI API",
		APIVersion:       version,
		Log:              log,
		Store:            dbcli,
		TokenAuthEnabled: apiToken != "",
	}, middlewares...)

	// Compute frontend source label.
	var frontendSource string
	if *flagServeFrontend {
		if *flagFrontendDir != "" {
			frontendSource = "directory"
		} else {
			frontendSource = "embedded"
		}
	}

	// Build meta provider config from flags (no secrets exposed)
	appConfig := meta.AppConfig{
		ListenAddress: *flagListen,
		DebugEnabled:  *flagDebug,
		Security: meta.SecurityConfig{
			TLSEnabled:       tlsResult != nil,
			TLSAutoGenerated: tlsResult != nil && tlsResult.AutoCert,
			MTLSEnabled:      tlsResult != nil && tlsResult.MTLSEnabled,
			TokenAuthEnabled: apiToken != "",
			RBACEnabled:      *flagEnableRBAC,
		},
		Frontend: meta.FrontendConfig{
			Enabled: *flagServeFrontend,
			Source:  frontendSource,
			Dir:     *flagFrontendDir,
		},
		Storage: meta.StorageConfig{
			DataDir: *flagDataDir,
		},
		Metrics: meta.MetricsConfig{
			Interval:  *flagMetricsInterval,
			Retention: *flagMetricsRetention,
		},
	}

	schemaVerFn := func() (int, error) {
		return dbcli.GetSchemaVersion()
	}

	// Populated after all providers are constructed; closure is only called at
	// request time so the slice is fully built by then.
	var allProviders []provider.Provider

	providersFn := func() []meta.ProviderStatus {
		type statusInfoer interface {
			StatusInfo() provider.StatusInfo
		}
		out := make([]meta.ProviderStatus, len(allProviders))
		for i, p := range allProviders {
			si := p.(statusInfoer).StatusInfo()
			out[i] = meta.ProviderStatus{
				Name:          p.Name(),
				Enabled:       si.Enabled,
				Running:       si.Running,
				EndpointCount: si.EndpointCount,
			}
		}
		return out
	}

	storeInfoFn := func(ctx context.Context) meta.StoreInfo {
		info := meta.StoreInfo{
			Engine: "sqlite",
			Path:   dbcli.Path(),
		}
		if fi, err := os.Stat(dbcli.Path()); err == nil {
			info.FileSizeBytes = fi.Size()
		}
		if v, err := dbcli.GetSchemaVersion(); err == nil {
			info.SchemaVersion = v
		}

		diagKey := store.Key{Namespace: "_diagnostics", ID: "healthcheck"}
		checks := []meta.DiagnosticCheck{
			runCheck("ping", func() error { return dbcli.Health(ctx) }),
			runCheck("write", func() error {
				_, err := store.Save(dbcli, ctx, diagKey, "ok")
				return err
			}),
			runCheck("read", func() error {
				item, ok, err := store.Load[string](dbcli, ctx, diagKey)
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
			runCheck("delete", func() error { return dbcli.Delete(ctx, diagKey) }),
		}
		info.Diagnostics = checks

		passed := 0
		for _, c := range checks {
			if c.Passed {
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

	metaProvider := meta.NewProvider(
		meta.VersionInfo{
			Version:    version,
			BuildDate:  buildDate,
			Branch:     branch,
			CommitHash: commitHash,
		},
		appConfig,
		schemaVerFn,
		providersFn,
		storeInfoFn,
		transport.ProviderOpts("meta")...,
	)

	allProviders = append(allProviders, metaProvider)

	// Serve frontend if enabled
	if *flagServeFrontend {
		var frontendHandler http.Handler
		if *flagFrontendDir != "" {
			log.Info("serving frontend from directory", "path", *flagFrontendDir)
			frontendHandler = frontend.NewHandler(os.DirFS(*flagFrontendDir), "")
		} else {
			log.Info("serving frontend from embedded files")
			frontendHandler = frontend.NewHandler(frontend.DistFS, "dist")
		}
		transport.Mount("/*", frontendHandler)
	}

	// Create server with explicit configuration
	server := &http.Server{
		Addr:    *flagListen,
		Handler: transport.Handler(),
	}

	// Start server in goroutine
	go func() {
		if tlsResult != nil {
			server.TLSConfig = tlsResult.Config
			log.Info("starting HTTPS server", "addr", *flagListen)
			if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Error("HTTPS server error", "error", err)
				os.Exit(1)
			}
		} else {
			log.Info("starting HTTP server", "addr", *flagListen)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error("HTTP server error", "error", err)
				os.Exit(1)
			}
		}
	}()

	// Signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var lastSignal time.Time
	confirmWindow := 3 * time.Second

	for {
		<-sigChan
		now := time.Now()
		if now.Sub(lastSignal) <= confirmWindow {
			// Second press within window - shutdown
			log.Info("shutting down server...")

			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdownCancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				log.Error("server shutdown error", "error", err)
			}
			return
		}
		// First press - warn user
		lastSignal = now
		log.Warn("interrupt received, press Ctrl+C again within 3 seconds to quit")
	}
}

func runCheck(name string, fn func() error) meta.DiagnosticCheck {
	start := time.Now()
	err := fn()
	d := time.Since(start)
	c := meta.DiagnosticCheck{
		Name:    name,
		Passed:  err == nil,
		Latency: d.String(),
	}
	if err != nil {
		c.Error = err.Error()
	}
	return c
}

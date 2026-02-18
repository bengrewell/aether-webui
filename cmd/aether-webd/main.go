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
	"github.com/bengrewell/aether-webui/internal/frontend"
	"github.com/bengrewell/aether-webui/internal/logging"
	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/provider/meta"
	"github.com/bengrewell/aether-webui/internal/sqlite"
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
	flagTLSCert := u.AddStringOption("t", "tls-cert", "", "Path to the TLS certificate file for HTTPS. When provided with --tls-key, the server will use HTTPS instead of HTTP.", "", secOptions)
	flagTLSKey := u.AddStringOption("k", "tls-key", "", "Path to the TLS private key file for HTTPS. Required when --tls-cert is specified.", "", secOptions)
	flagMTLSCACert := u.AddStringOption("m", "mtls-ca-cert", "", "Path to the CA certificate file for verifying client certificates. Enables mutual TLS (mTLS) authentication when specified.", "", secOptions)
	flagEnableRBAC := u.AddBooleanOption("r", "enable-rbac", false, "Enable role-based access control (RBAC) authentication and authorization middleware.", "", secOptions)

	exeOptions := u.AddGroup(1, "Execution Options", "Options that control API command execution")
	flagExecUser := u.AddStringOption("u", "exec-user", "", "User account under which API commands will be executed", "", exeOptions)
	flagExecEnv := u.AddStringOption("e", "exec-env", "", "Environment variables to pass to the command execution context", "", exeOptions)

	frontendOptions := u.AddGroup(3, "Frontend Options", "Options that control frontend serving")
	flagServeFrontend := u.AddBooleanOption("f", "serve-frontend", true, "Enable serving frontend static files from embedded or custom directory", "", frontendOptions)
	flagFrontendDir := u.AddStringOption("", "frontend-dir", "", "Override embedded frontend with files from this directory (for development)", "", frontendOptions)

	storageOptions := u.AddGroup(4, "Storage Options", "Options that control persistent state storage")
	flagDataDir := u.AddStringOption("", "data-dir", "/var/lib/aether-webd", "Directory for persistent state database", "", storageOptions)
	flagOnRampDir := u.AddStringOption("", "onramp-dir", "", "Directory for OnRamp repository checkout. Defaults to <data-dir>/onramp", "", storageOptions)
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

	// Log unimplemented flag values at debug level
	log.Debug("unimplemented security options",
		"tls_cert", *flagTLSCert,
		"tls_key", *flagTLSKey,
		"mtls_ca_cert", *flagMTLSCACert,
		"enable_rbac", *flagEnableRBAC,
	)
	log.Debug("unimplemented execution options",
		"exec_user", *flagExecUser,
		"exec_env", *flagExecEnv,
	)

	// Open database using --data-dir flag
	dbPath := filepath.Join(*flagDataDir, "app.db")
	if err := os.MkdirAll(*flagDataDir, 0o750); err != nil {
		log.Error("failed to create database directory", "path", *flagDataDir, "err", err)
		os.Exit(1)
	}

	st, err := sqlite.Open(context.Background(), sqlite.Config{
		Path:          dbPath,
		BusyTimeout:   5 * time.Second,
		Crypter:       sqlite.NoopCrypter{},
		MetricsMaxAge: 7 * 24 * time.Hour,
	})
	if err != nil {
		log.Error("sqlite open failed", "path", dbPath, "err", err)
		os.Exit(1)
	}

	if err := st.Migrate(context.Background()); err != nil {
		log.Error("sqlite migrate failed", "err", err)
	}

	dbcli := store.Client{S: st, C: store.JSONCodec{}}

	// Create REST transport (Chi router + Huma API + shared deps)
	transport := rest.NewTransport(rest.Config{
		APITitle:   "Aether WebUI API",
		APIVersion: version,
		Log:        log,
		Store:      dbcli,
	}, logging.RequestLogger())

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
			TLSEnabled:  *flagTLSCert != "" && *flagTLSKey != "",
			MTLSEnabled: *flagMTLSCACert != "",
			RBACEnabled: *flagEnableRBAC,
		},
		Frontend: meta.FrontendConfig{
			Enabled: *flagServeFrontend,
			Source:  frontendSource,
			Dir:     *flagFrontendDir,
		},
		Storage: meta.StorageConfig{
			DataDir:   *flagDataDir,
			OnRampDir: *flagOnRampDir,
		},
		Metrics: meta.MetricsConfig{
			Interval:  *flagMetricsInterval,
			Retention: *flagMetricsRetention,
		},
	}

	schemaVerFn := func() (int, error) {
		return dbcli.S.GetSchemaVersion()
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
		log.Info("starting HTTP server", "addr", *flagListen)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server error", "error", err)
			os.Exit(1)
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

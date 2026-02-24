package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bengrewell/aether-webui/internal/controller"
	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/provider/meta"
	"github.com/bengrewell/aether-webui/internal/provider/nodes"
	"github.com/bengrewell/aether-webui/internal/provider/onramp"
	"github.com/bengrewell/aether-webui/internal/provider/system"
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
	_ = u.AddStringOption("u", "exec-user", "", "User account under which API commands will be executed", "", exeOptions)
	_ = u.AddStringOption("e", "exec-env", "", "Environment variables to pass to the command execution context", "", exeOptions)

	onrampOptions := u.AddGroup(6, "OnRamp Options", "Options that control the Aether OnRamp provider")
	flagOnRampDir := u.AddStringOption("", "onramp-dir", "", "Path to aether-onramp repo (default: {data-dir}/aether-onramp)", "", onrampOptions)
	flagOnRampVersion := u.AddStringOption("", "onramp-version", "main", "Tag, branch, or commit to pin aether-onramp to", "", onrampOptions)

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

	if !parsed {
		u.PrintUsage()
		return
	}

	if *flagVersion {
		fmt.Printf("aether-webd %s (branch: %s, commit: %s, built: %s)\n",
			version, branch, commitHash, buildDate)
		return
	}

	collectInterval, err := time.ParseDuration(*flagMetricsInterval)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --metrics-interval: %v\n", err)
		os.Exit(1)
	}

	ctrl, err := controller.New(
		controller.WithVersion(meta.VersionInfo{
			Version:    version,
			BuildDate:  buildDate,
			Branch:     branch,
			CommitHash: commitHash,
		}),
		controller.WithListenAddr(*flagListen),
		controller.WithDebug(*flagDebug),
		controller.WithDataDir(*flagDataDir),
		controller.WithTLS(*flagTLS, *flagTLSCert, *flagTLSKey, *flagMTLSCACert),
		controller.WithAPIToken(*flagAPIToken),
		controller.WithRBAC(*flagEnableRBAC),
		controller.WithFrontend(*flagServeFrontend, *flagFrontendDir),
		controller.WithMetrics(*flagMetricsInterval, *flagMetricsRetention),
		controller.WithEncryptionKey(*flagEncryptionKey),
		controller.WithProvider("system", true, func(_ context.Context, _ store.Client, opts []provider.Option) (provider.Provider, error) {
			return system.NewProvider(system.Config{
				CollectInterval: collectInterval,
			}, opts...), nil
		}),
		controller.WithProvider("nodes", true, func(_ context.Context, _ store.Client, opts []provider.Option) (provider.Provider, error) {
			return nodes.NewProvider(opts...), nil
		}),
		controller.WithProvider("onramp", true, func(_ context.Context, _ store.Client, opts []provider.Option) (provider.Provider, error) {
			dir := *flagOnRampDir
			if dir == "" {
				dir = filepath.Join(*flagDataDir, "aether-onramp")
			}
			return onramp.NewProvider(onramp.Config{
				OnRampDir: dir,
				RepoURL:   "https://github.com/opennetworkinglab/aether-onramp.git",
				Version:   *flagOnRampVersion,
			}, opts...), nil
		}),
	)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if err := ctrl.Run(context.Background()); err != nil {
		os.Exit(1)
	}
}

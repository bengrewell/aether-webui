package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bengrewell/aether-webui/internal/controller"
	"github.com/bengrewell/aether-webui/internal/nodefacts"
	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/provider/configdefaults"
	"github.com/bengrewell/aether-webui/internal/provider/meta"
	"github.com/bengrewell/aether-webui/internal/provider/nodes"
	"github.com/bengrewell/aether-webui/internal/provider/onramp"
	"github.com/bengrewell/aether-webui/internal/provider/preflight"
	"github.com/bengrewell/aether-webui/internal/provider/system"
	"github.com/bengrewell/aether-webui/internal/provider/wizard"
	"github.com/bengrewell/aether-webui/internal/store"
	"github.com/bgrewell/usage"
)

var (
	version    string = "dev"
	buildDate  string = "unknown"
	branch     string = "unknown"
	commitHash string = "unknown"
)

// envOr returns the value of the named environment variable, or fallback if
// the variable is empty or unset.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envBool returns the boolean value of the named environment variable, or
// fallback if the variable is empty, unset, or not a recognised boolean.
func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	switch v {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return fallback
	}
}

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
	flagDebug := u.AddBooleanOption("d", "debug", envBool("AETHER_DEBUG", false), "Enable debug mode for verbose logging and diagnostic output (env: AETHER_DEBUG)", "", nil)
	flagListen := u.AddStringOption("l", "listen", envOr("AETHER_LISTEN", "127.0.0.1:8186"), "Address and port the API server will listen on (env: AETHER_LISTEN)", "", nil)

	secOptions := u.AddGroup(2, "Security Options", "Options that control security settings")
	flagTLS := u.AddBooleanOption("", "tls", envBool("AETHER_TLS", false), "Enable TLS; auto-generates a self-signed certificate if --tls-cert and --tls-key are not provided (env: AETHER_TLS)", "", secOptions)
	flagTLSCert := u.AddStringOption("t", "tls-cert", envOr("AETHER_TLS_CERT", ""), "Path to the TLS certificate file for HTTPS (env: AETHER_TLS_CERT)", "", secOptions)
	flagTLSKey := u.AddStringOption("k", "tls-key", envOr("AETHER_TLS_KEY", ""), "Path to the TLS private key file for HTTPS (env: AETHER_TLS_KEY)", "", secOptions)
	flagMTLSCACert := u.AddStringOption("m", "mtls-ca-cert", envOr("AETHER_MTLS_CA_CERT", ""), "Path to the CA certificate file for client verification; enables mTLS (env: AETHER_MTLS_CA_CERT)", "", secOptions)
	flagAPIToken := u.AddStringOption("", "api-token", envOr("AETHER_API_TOKEN", ""), "Bearer token for API authentication; all /api/* requests require Authorization: Bearer <token> (env: AETHER_API_TOKEN)", "", secOptions)
	flagEnableRBAC := u.AddBooleanOption("r", "enable-rbac", envBool("AETHER_ENABLE_RBAC", false), "Enable role-based access control (RBAC) authentication and authorization middleware (env: AETHER_ENABLE_RBAC)", "", secOptions)
	flagCORSOrigins := u.AddStringOption("", "cors-origins", envOr("AETHER_CORS_ORIGINS", ""), "Comma-separated list of allowed CORS origins, e.g. http://localhost:5173 (env: AETHER_CORS_ORIGINS)", "", secOptions)

	exeOptions := u.AddGroup(1, "Execution Options", "Options that control API command execution")
	_ = u.AddStringOption("u", "exec-user", envOr("AETHER_EXEC_USER", ""), "User account under which API commands will be executed (env: AETHER_EXEC_USER)", "", exeOptions)
	_ = u.AddStringOption("e", "exec-env", envOr("AETHER_EXEC_ENV", ""), "Environment variables to pass to the command execution context (env: AETHER_EXEC_ENV)", "", exeOptions)

	onrampOptions := u.AddGroup(6, "OnRamp Options", "Options that control the Aether OnRamp provider")
	flagOnRampDir := u.AddStringOption("", "onramp-dir", envOr("AETHER_ONRAMP_DIR", ""), "Path to aether-onramp repo; default: {data-dir}/aether-onramp (env: AETHER_ONRAMP_DIR)", "", onrampOptions)
	flagOnRampVersion := u.AddStringOption("", "onramp-version", envOr("AETHER_ONRAMP_VERSION", "main"), "Tag, branch, or commit to pin aether-onramp to (env: AETHER_ONRAMP_VERSION)", "", onrampOptions)

	frontendOptions := u.AddGroup(3, "Frontend Options", "Options that control frontend serving")
	flagServeFrontend := u.AddBooleanOption("f", "serve-frontend", envBool("AETHER_SERVE_FRONTEND", true), "Enable serving frontend static files from embedded or custom directory (env: AETHER_SERVE_FRONTEND)", "", frontendOptions)
	flagFrontendDir := u.AddStringOption("", "frontend-dir", envOr("AETHER_FRONTEND_DIR", ""), "Override embedded frontend with files from this directory (env: AETHER_FRONTEND_DIR)", "", frontendOptions)

	storageOptions := u.AddGroup(4, "Storage Options", "Options that control persistent state storage")
	flagDataDir := u.AddStringOption("", "data-dir", envOr("AETHER_DATA_DIR", "/var/lib/aether-webd"), "Directory for persistent state database (env: AETHER_DATA_DIR)", "", storageOptions)
	flagEncryptionKey := u.AddStringOption("", "encryption-key", envOr("AETHER_ENCRYPTION_KEY", ""), "32-byte encryption key for node passwords; auto-generated if not provided (env: AETHER_ENCRYPTION_KEY)", "", secOptions)

	mcpOptions := u.AddGroup(7, "MCP Options", "Options for the embedded MCP server")
	flagMCP := u.AddBooleanOption("", "mcp", envBool("AETHER_MCP", false), "Enable MCP server for LLM tool integration via stdio (env: AETHER_MCP)", "", mcpOptions)
	flagMCPListen := u.AddStringOption("", "mcp-listen", envOr("AETHER_MCP_LISTEN", ""), "Address for MCP StreamableHTTP transport; enables HTTP-based MCP (env: AETHER_MCP_LISTEN)", "", mcpOptions)

	metricsOptions := u.AddGroup(5, "Metrics Options", "Options that control metrics collection")
	flagMetricsInterval := u.AddStringOption("", "metrics-interval", envOr("AETHER_METRICS_INTERVAL", "10s"), "How often to collect system metrics, e.g. 10s, 30s, 1m (env: AETHER_METRICS_INTERVAL)", "", metricsOptions)
	flagMetricsRetention := u.AddStringOption("", "metrics-retention", envOr("AETHER_METRICS_RETENTION", "24h"), "How long to retain historical metrics data, e.g. 24h, 7d (env: AETHER_METRICS_RETENTION)", "", metricsOptions)

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

	var corsOrigins []string
	if *flagCORSOrigins != "" {
		for _, o := range strings.Split(*flagCORSOrigins, ",") {
			if o = strings.TrimSpace(o); o != "" {
				corsOrigins = append(corsOrigins, o)
			}
		}
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
		controller.WithCORSOrigins(corsOrigins),
		controller.WithFrontend(*flagServeFrontend, *flagFrontendDir),
		controller.WithMetrics(*flagMetricsInterval, *flagMetricsRetention),
		controller.WithEncryptionKey(*flagEncryptionKey),
		controller.WithMCP(*flagMCP || *flagMCPListen != ""),
		controller.WithMCPListenAddr(*flagMCPListen),
		controller.WithProvider("system", true, func(_ context.Context, _ store.Client, opts []provider.Option) (provider.Provider, error) {
			return system.NewProvider(system.Config{
				CollectInterval: collectInterval,
			}, opts...), nil
		}),
		controller.WithProvider("nodes", true, func(_ context.Context, _ store.Client, opts []provider.Option) (provider.Provider, error) {
			return nodes.NewProvider(opts...), nil
		}),
		controller.WithProvider("preflight", true, func(_ context.Context, _ store.Client, opts []provider.Option) (provider.Provider, error) {
			return preflight.NewProvider(opts...), nil
		}),
		controller.WithProvider("wizard", true, func(_ context.Context, _ store.Client, opts []provider.Option) (provider.Provider, error) {
			return wizard.NewProvider(opts...), nil
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
		controller.WithProvider("configdefaults", true, func(_ context.Context, _ store.Client, opts []provider.Option) (provider.Provider, error) {
			dir := *flagOnRampDir
			if dir == "" {
				dir = filepath.Join(*flagDataDir, "aether-onramp")
			}
			return configdefaults.NewProvider(configdefaults.Config{
				OnRampDir: dir,
			}, &nodefacts.SSHGatherer{}, opts...), nil
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

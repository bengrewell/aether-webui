package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bengrewell/aether-webui/internal/aether"
	"github.com/bengrewell/aether-webui/internal/frontend"
	"github.com/bengrewell/aether-webui/internal/k8sinfo"
	"github.com/bengrewell/aether-webui/internal/state"
	"github.com/bengrewell/aether-webui/internal/sysinfo"
	"github.com/bengrewell/aether-webui/internal/webuiapi"
	"github.com/bgrewell/usage"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

var (
	version    string = "dev"
	buildDate  string = "unknown"
	branch     string = "unknown"
	commitHash string = "unknown"
)

func main() {

	u := usage.NewUsage(
		usage.WithApplicationName("aether-webd"),
		usage.WithApplicationVersion(version),
		usage.WithApplicationBuildDate(buildDate),
		usage.WithApplicationCommitHash(commitHash),
		usage.WithApplicationBranch(branch),
		usage.WithApplicationDescription("Backend API service for the Aether WebUI. This service is responsible for executing deployment tasks, gathering system information, and monitoring the health and metrics of Aether 5G deployments. It manages SD-Core components, gNBs (such as srsRAN and OCUDU), Kubernetes clusters, and host systems."),
	)

	flagDebug := u.AddBooleanOption("d", "debug", false, "Enable debug mode for verbose logging and diagnostic output", "", nil)
	flagListen := u.AddStringOption("l", "listen", "127.0.0.1:8680", "Address and port the API server will listen on (e.g., 0.0.0.0:8680 for all interfaces)", "", nil)

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

	parsed := u.Parse()

	if !parsed {
		u.PrintUsage()
		return
	}

	_ = flagDebug
	_ = flagTLSCert
	_ = flagTLSKey
	_ = flagMTLSCACert
	_ = flagEnableRBAC
	_ = flagExecUser
	_ = flagExecEnv

	// Initialize persistent state store
	stateStore, err := state.NewSQLiteStore(*flagDataDir)
	if err != nil {
		fmt.Printf("Failed to initialize state store: %v\n", err)
		os.Exit(1)
	}
	defer stateStore.Close()

	// Create Chi router and Huma API
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Aether WebUI API", version))

	// Initialize providers with mock implementations
	sysProvider := sysinfo.NewMockProvider()
	sysResolver := sysinfo.NewDefaultNodeResolver(sysProvider)
	k8sProvider := k8sinfo.NewMockProvider()
	aetherProvider := aether.NewMockProvider("local")
	aetherResolver := aether.NewDefaultHostResolver(aetherProvider)

	// Register routes
	webuiapi.RegisterHealthRoutes(api)
	webuiapi.RegisterSetupRoutes(api, stateStore)
	webuiapi.RegisterSystemRoutes(api, sysResolver)
	webuiapi.RegisterMetricsRoutes(api, sysResolver)
	webuiapi.RegisterKubernetesRoutes(api, k8sProvider)
	webuiapi.RegisterAetherRoutes(api, aetherResolver)

	// Serve frontend if enabled
	if *flagServeFrontend {
		var frontendHandler http.Handler
		if *flagFrontendDir != "" {
			// Serve from custom directory
			fmt.Printf("Serving frontend from directory: %s\n", *flagFrontendDir)
			frontendHandler = frontend.NewHandler(os.DirFS(*flagFrontendDir), "")
		} else {
			// Serve from embedded files
			fmt.Println("Serving frontend from embedded files")
			frontendHandler = frontend.NewHandler(frontend.DistFS, "dist")
		}
		// Mount frontend handler as catch-all (after API routes)
		router.Handle("/*", frontendHandler)
	}

	// Start HTTP server
	fmt.Printf("Starting server on %s\n", *flagListen)
	if err := http.ListenAndServe(*flagListen, router); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

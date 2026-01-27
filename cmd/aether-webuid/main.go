package main

import (
	"github.com/bgrewell/usage"
)

var (
	version    string = "dev"
	buildDate  string = "unknown"
	branch     string = "unknown"
	commitHash string = "unknown"
)

func main() {

	u := usage.NewUsage(
		usage.WithApplicationName("aether-webuid"),
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

	parsed := u.Parse()

	if !parsed {
		u.PrintUsage()
	}

	_ = flagDebug
	_ = flagListen
	_ = flagTLSCert
	_ = flagTLSKey
	_ = flagMTLSCACert
	_ = flagEnableRBAC
	_ = flagExecUser
	_ = flagExecEnv

}

package main

import (
	"time"

	"github.com/bgrewell/usage"
)

var (
	version    string = "debug"
	buildDate  string = time.Now().Format("2006-01-02 15:04:05")
	branch     string = "debug"
	commitHash string = "debug"
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

	secOptions := u.AddGroup(2, "Security Options", "Options that control security settings")
	flagSkipTLSVerify := u.AddBooleanOption("k", "skip-tls-verify", false, "Skip TLS certificate verification when connecting to remote services", "", secOptions)
	flagTLSCACert := u.AddStringOption("c", "tls-ca-cert", "", "Path to a custom CA certificate file for TLS verification", "", secOptions)
	flagTLSClientCert := u.AddStringOption("C", "tls-client-cert", "", "Path to the client TLS certificate file for mutual TLS authentication", "", secOptions)
	flagTLSClientKey := u.AddStringOption("K", "tls-client-key", "", "Path to the client TLS private key file for mutual TLS authentication", "", secOptions)

	exeOptions := u.AddGroup(1, "Execution Options", "Options that control API command execution")
	flagExecUser := u.AddStringOption("u", "exec-user", "", "User account under which API commands will be executed", "", exeOptions)
	flagExecEnv := u.AddStringOption("e", "exec-env", "", "Environment variables to pass to the command execution context", "", exeOptions)

	parsed := u.Parse()

	if !parsed {
		u.PrintUsage()
	}

	_ = flagDebug
	_ = flagSkipTLSVerify
	_ = flagTLSCACert
	_ = flagTLSClientCert
	_ = flagTLSClientKey
	_ = flagExecUser
	_ = flagExecEnv

}

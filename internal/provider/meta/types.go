package meta

// VersionInfo holds build-time version metadata.
type VersionInfo struct {
	Version    string `json:"version"`
	BuildDate  string `json:"buildDate"`
	Branch     string `json:"branch"`
	CommitHash string `json:"commitHash"`
}

// VersionOutput is the Huma response wrapper for VersionInfo.
type VersionOutput struct {
	Body VersionInfo
}

// BuildInfo holds Go toolchain and target platform details.
type BuildInfo struct {
	GoVersion string `json:"goVersion"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// BuildOutput is the Huma response wrapper for BuildInfo.
type BuildOutput struct {
	Body BuildInfo
}

// UserInfo holds the uid/name of the process owner.
type UserInfo struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

// GroupInfo holds the gid/name of the process primary group.
type GroupInfo struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

// RuntimeInfo holds process-level runtime details.
type RuntimeInfo struct {
	PID        int       `json:"pid"`
	User       UserInfo  `json:"user"`
	Group      GroupInfo `json:"group"`
	BinaryPath string   `json:"binaryPath"`
	StartTime  string   `json:"startTime"`
	Uptime     string   `json:"uptime"`
}

// RuntimeOutput is the Huma response wrapper for RuntimeInfo.
type RuntimeOutput struct {
	Body RuntimeInfo
}

// FrontendConfig describes how the frontend is being served.
type FrontendConfig struct {
	Enabled bool   `json:"enabled"`
	Source  string `json:"source"`
	Dir     string `json:"dir"`
}

// AppConfig holds non-secret application configuration values for introspection.
type AppConfig struct {
	ListenAddress    string         `json:"listenAddress"`
	DataDir          string         `json:"dataDir"`
	TLSEnabled       bool           `json:"tlsEnabled"`
	MTLSEnabled      bool           `json:"mtlsEnabled"`
	RBACEnabled      bool           `json:"rbacEnabled"`
	DebugEnabled     bool           `json:"debugEnabled"`
	Frontend         FrontendConfig `json:"frontend"`
	OnRampDir        string         `json:"onrampDir"`
	MetricsInterval  string         `json:"metricsInterval"`
	MetricsRetention string         `json:"metricsRetention"`
}

// ConfigInfo holds the active application configuration plus schema version.
type ConfigInfo struct {
	AppConfig
	SchemaVersion int `json:"schemaVersion"`
}

// ConfigOutput is the Huma response wrapper for ConfigInfo.
type ConfigOutput struct {
	Body ConfigInfo
}

// ProviderStatus describes a single registered provider's state.
type ProviderStatus struct {
	Name          string `json:"name"`
	Enabled       bool   `json:"enabled"`
	Running       bool   `json:"running"`
	EndpointCount int    `json:"endpointCount"`
}

// ProvidersInfo holds the list of registered provider statuses.
type ProvidersInfo struct {
	Providers []ProviderStatus `json:"providers"`
}

// ProvidersOutput is the Huma response wrapper for ProvidersInfo.
type ProvidersOutput struct {
	Body ProvidersInfo
}

// SchemaVersionFunc returns the current database schema version.
type SchemaVersionFunc func() (int, error)

// ProviderStatusFunc returns the status of all registered providers.
type ProviderStatusFunc func() []ProviderStatus

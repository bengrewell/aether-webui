package mcp

// MCP tool input types. These are lightweight structs with json and jsonschema
// tags for automatic schema generation by the go-sdk. They are decoupled from
// Huma-specific input types.

// --- Nodes ---

type NodesListInput struct{}

type NodesGetInput struct {
	ID string `json:"id" jsonschema:"node ID"`
}

type NodesCreateInput struct {
	Name         string   `json:"name" jsonschema:"unique node name (Ansible inventory hostname)"`
	AnsibleHost  string   `json:"ansible_host" jsonschema:"IP or hostname for SSH"`
	AnsibleUser  string   `json:"ansible_user" jsonschema:"SSH username"`
	Password     string   `json:"password" jsonschema:"SSH password"`
	SudoPassword string   `json:"sudo_password" jsonschema:"sudo password"`
	SSHKey       string   `json:"ssh_key,omitempty" jsonschema:"SSH private key (optional)"`
	Roles        []string `json:"roles,omitempty" jsonschema:"role assignments (e.g. master, worker, gnbsim)"`
}

type NodesUpdateInput struct {
	ID           string   `json:"id" jsonschema:"node ID"`
	Name         *string  `json:"name,omitempty" jsonschema:"unique node name"`
	AnsibleHost  *string  `json:"ansible_host,omitempty" jsonschema:"IP or hostname for SSH"`
	AnsibleUser  *string  `json:"ansible_user,omitempty" jsonschema:"SSH username"`
	Password     *string  `json:"password,omitempty" jsonschema:"SSH password"`
	SudoPassword *string  `json:"sudo_password,omitempty" jsonschema:"sudo password"`
	SSHKey       *string  `json:"ssh_key,omitempty" jsonschema:"SSH private key"`
	Roles        []string `json:"roles,omitempty" jsonschema:"role assignments (replaces entire set)"`
}

type NodesDeleteInput struct {
	ID string `json:"id" jsonschema:"node ID"`
}

// --- OnRamp ---

type ComponentsListInput struct{}

type ComponentGetInput struct {
	Component string `json:"component" jsonschema:"component name (e.g. k8s, 5gc, gnbsim)"`
}

type DeployActionInput struct {
	Component string            `json:"component" jsonschema:"component name"`
	Action    string            `json:"action" jsonschema:"action name (e.g. install, uninstall)"`
	Labels    map[string]string `json:"labels,omitempty" jsonschema:"optional labels for the action"`
	Tags      []string          `json:"tags,omitempty" jsonschema:"optional tags for the action"`
}

type RepoStatusInput struct{}

type RepoRefreshInput struct{}

type ConfigGetInput struct{}

type ConfigPatchInput struct {
	Config map[string]any `json:"config" jsonschema:"partial config to merge into vars/main.yml"`
}

type ProfilesListInput struct{}

// --- Tasks ---

type TasksListInput struct{}

type TaskGetInput struct {
	ID     string `json:"id" jsonschema:"task ID"`
	Offset int    `json:"offset,omitempty" jsonschema:"byte offset for incremental output reads"`
}

type TaskCancelInput struct {
	ID string `json:"id" jsonschema:"task ID to cancel"`
}

type ActionsListInput struct {
	Component string `json:"component,omitempty" jsonschema:"filter by component name"`
	Action    string `json:"action,omitempty" jsonschema:"filter by action name"`
	Status    string `json:"status,omitempty" jsonschema:"filter by status"`
	Limit     int    `json:"limit,omitempty" jsonschema:"max results (default 50)"`
	Offset    int    `json:"offset,omitempty" jsonschema:"pagination offset"`
}

type ActionGetInput struct {
	ID string `json:"id" jsonschema:"action ID"`
}

// --- System ---

type SystemOverviewInput struct{}

type SystemNetworkInput struct{}

type SystemMetricsInput struct {
	Metric      string `json:"metric" jsonschema:"metric name to query (e.g. system.cpu.usage_percent)"`
	From        string `json:"from,omitempty" jsonschema:"start time (RFC 3339)"`
	To          string `json:"to,omitempty" jsonschema:"end time (RFC 3339)"`
	Labels      string `json:"labels,omitempty" jsonschema:"comma-separated key=val label filters"`
	Aggregation string `json:"aggregation,omitempty" jsonschema:"time bucket aggregation: raw, 10s, 1m, 5m, 1h"`
}

// --- Meta ---

type ServerStatusInput struct{}

type ComponentStatesListInput struct{}

type ComponentStateGetInput struct {
	Component string `json:"component" jsonschema:"component name"`
}

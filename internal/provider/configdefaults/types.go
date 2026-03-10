package configdefaults

import (
	"github.com/bengrewell/aether-webui/internal/nodefacts"
	"github.com/bengrewell/aether-webui/internal/provider/onramp"
)

// --- Node facts endpoint ---

// NodeFactsGetInput is the input for the get-node-facts endpoint.
type NodeFactsGetInput struct {
	ID      string `path:"id" doc:"Node ID"`
	Refresh bool   `query:"refresh" default:"false" doc:"Force SSH re-gathering of facts"`
}

// NodeFactsGetOutput wraps the response for the get-node-facts endpoint.
type NodeFactsGetOutput struct {
	Body nodefacts.NodeFacts
}

// --- Config defaults endpoint ---

// ConfigDefaultsApplyInput is the input for the apply-config-defaults endpoint.
type ConfigDefaultsApplyInput struct {
	Refresh bool `query:"refresh" default:"false" doc:"Force SSH re-gathering of facts for all nodes"`
}

// AppliedDefault describes a single config field that was set by the defaults engine.
type AppliedDefault struct {
	Field       string `json:"field"`
	Value       any    `json:"value"`
	Explanation string `json:"explanation"`
	SourceNode  string `json:"source_node"`
}

// ConfigDefaultsApplyOutput wraps the response for the apply-config-defaults endpoint.
type ConfigDefaultsApplyOutput struct {
	Body ConfigDefaultsResult
}

// ConfigDefaultsResult is the body returned by the apply-config-defaults endpoint.
type ConfigDefaultsResult struct {
	Applied []AppliedDefault `json:"applied"`
	Errors  []string         `json:"errors"`
	Config  onramp.OnRampConfig `json:"config"`
}

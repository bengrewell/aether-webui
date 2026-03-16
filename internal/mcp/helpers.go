package mcp

import (
	"encoding/json"
	"fmt"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// jsonResult marshals v to JSON and returns it as a TextContent CallToolResult.
func jsonResult(v any) *gomcp.CallToolResult {
	data, err := json.Marshal(v)
	if err != nil {
		return errorResult(fmt.Errorf("marshal result: %w", err))
	}
	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}
}

// errorResult returns a CallToolResult with IsError set.
func errorResult(err error) *gomcp.CallToolResult {
	r := &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: err.Error()}},
		IsError: true,
	}
	return r
}

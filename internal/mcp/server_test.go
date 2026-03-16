package mcp

import (
	"encoding/json"
	"testing"
	"time"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/provider/meta"
	"github.com/bengrewell/aether-webui/internal/provider/nodes"
	"github.com/bengrewell/aether-webui/internal/provider/onramp"
	"github.com/bengrewell/aether-webui/internal/provider/system"
	"github.com/bengrewell/aether-webui/internal/store"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	ctx := t.Context()
	dbPath := t.TempDir() + "/test.db"
	st, err := store.New(ctx, dbPath)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	n := nodes.NewProvider(provider.WithStore(st))
	o := onramp.NewProvider(onramp.Config{
		OnRampDir: t.TempDir(),
		RepoURL:   "https://example.com/repo.git",
		Version:   "main",
	}, provider.WithStore(st))
	s := system.NewProvider(system.Config{
		CollectInterval: 10 * time.Second,
	}, provider.WithStore(st))
	m := meta.NewProvider(
		meta.VersionInfo{Version: "test", BuildDate: "now"},
		meta.AppConfig{},
		func() (int, error) { return 1, nil },
		func() []meta.ProviderStatus { return nil },
		nil,
	)

	return New(Config{
		Store:   st,
		Nodes:   n,
		OnRamp:  o,
		System:  s,
		Meta:    m,
		Version: "test",
	})
}

// newTestSession creates an in-process MCP client session connected to the server.
func newTestSession(t *testing.T, srv *Server) *gomcp.ClientSession {
	t.Helper()
	ctx := t.Context()

	clientTransport, serverTransport := gomcp.NewInMemoryTransports()

	_, err := srv.MCPServer().Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}

	client := gomcp.NewClient(&gomcp.Implementation{Name: "test-client"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	return session
}

// listToolNames uses an in-process MCP client to enumerate registered tools.
func listToolNames(t *testing.T, srv *Server) []string {
	t.Helper()
	session := newTestSession(t, srv)

	result, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	names := make([]string, len(result.Tools))
	for i, tool := range result.Tools {
		names[i] = tool.Name
	}
	return names
}

func TestNew_RegistersExpectedToolCount(t *testing.T) {
	srv := newTestServer(t)
	names := listToolNames(t, srv)

	// Expected: 5 nodes + 8 onramp + 5 tasks + 3 system + 3 meta = 24 tools
	const expectedCount = 24
	if len(names) != expectedCount {
		t.Errorf("got %d tools, want %d\ntools: %v", len(names), expectedCount, names)
	}
}

func TestNew_ToolNamesPresent(t *testing.T) {
	srv := newTestServer(t)
	names := listToolNames(t, srv)

	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}

	required := []string{
		"nodes_list", "nodes_get", "nodes_create", "nodes_update", "nodes_delete",
		"components_list", "component_get", "deploy_action", "repo_status",
		"repo_refresh", "config_get", "config_patch", "profiles_list",
		"tasks_list", "task_get", "task_cancel", "actions_list", "action_get",
		"component_states_list", "component_state_get",
		"system_overview", "system_network", "system_metrics",
		"server_status",
	}

	for _, name := range required {
		if !nameSet[name] {
			t.Errorf("missing required tool: %s", name)
		}
	}
}

func TestCallTool_NodesList(t *testing.T) {
	srv := newTestServer(t)
	session := newTestSession(t, srv)

	result, err := session.CallTool(t.Context(), &gomcp.CallToolParams{
		Name: "nodes_list",
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}

	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	text, ok := result.Content[0].(*gomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var nodesList []any
	if err := json.Unmarshal([]byte(text.Text), &nodesList); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(nodesList) != 0 {
		t.Errorf("expected empty nodes list, got %d items", len(nodesList))
	}
}

func TestCallTool_ComponentsList(t *testing.T) {
	srv := newTestServer(t)
	session := newTestSession(t, srv)

	result, err := session.CallTool(t.Context(), &gomcp.CallToolParams{
		Name: "components_list",
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}

	text, ok := result.Content[0].(*gomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var components []map[string]any
	if err := json.Unmarshal([]byte(text.Text), &components); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(components) == 0 {
		t.Error("expected non-empty components list")
	}
}

func TestCallTool_ServerStatus(t *testing.T) {
	srv := newTestServer(t)
	session := newTestSession(t, srv)

	result, err := session.CallTool(t.Context(), &gomcp.CallToolParams{
		Name: "server_status",
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}

	text, ok := result.Content[0].(*gomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var status map[string]any
	if err := json.Unmarshal([]byte(text.Text), &status); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if status["version"] == nil {
		t.Error("expected version in server status")
	}
}

func TestCallTool_NodesCreateAndGet(t *testing.T) {
	srv := newTestServer(t)
	session := newTestSession(t, srv)
	ctx := t.Context()

	// Create a node.
	createResult, err := session.CallTool(ctx, &gomcp.CallToolParams{
		Name: "nodes_create",
		Arguments: NodesCreateInput{
			Name:         "test-node",
			AnsibleHost:  "192.168.1.100",
			AnsibleUser:  "admin",
			Password:     "secret",
			SudoPassword: "sudosecret",
			Roles:        []string{"master"},
		},
	})
	if err != nil {
		t.Fatalf("CallTool nodes_create: %v", err)
	}
	if createResult.IsError {
		t.Fatalf("nodes_create returned error: %v", createResult.Content)
	}

	// Extract the created node's ID.
	text := createResult.Content[0].(*gomcp.TextContent)
	var created map[string]any
	if err := json.Unmarshal([]byte(text.Text), &created); err != nil {
		t.Fatalf("unmarshal created node: %v", err)
	}
	nodeID, ok := created["id"].(string)
	if !ok || nodeID == "" {
		t.Fatal("expected non-empty node ID")
	}

	// Get the node by ID.
	getResult, err := session.CallTool(ctx, &gomcp.CallToolParams{
		Name:      "nodes_get",
		Arguments: NodesGetInput{ID: nodeID},
	})
	if err != nil {
		t.Fatalf("CallTool nodes_get: %v", err)
	}
	if getResult.IsError {
		t.Fatalf("nodes_get returned error: %v", getResult.Content)
	}

	getText := getResult.Content[0].(*gomcp.TextContent)
	var fetched map[string]any
	if err := json.Unmarshal([]byte(getText.Text), &fetched); err != nil {
		t.Fatalf("unmarshal fetched node: %v", err)
	}
	if fetched["name"] != "test-node" {
		t.Errorf("got name %q, want %q", fetched["name"], "test-node")
	}
	if fetched["ansible_host"] != "192.168.1.100" {
		t.Errorf("got ansible_host %q, want %q", fetched["ansible_host"], "192.168.1.100")
	}
}

func TestCallTool_TasksList(t *testing.T) {
	srv := newTestServer(t)
	session := newTestSession(t, srv)

	result, err := session.CallTool(t.Context(), &gomcp.CallToolParams{
		Name: "tasks_list",
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}

	text, ok := result.Content[0].(*gomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var tasks []any
	if err := json.Unmarshal([]byte(text.Text), &tasks); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	// Should be empty initially.
	if len(tasks) != 0 {
		t.Errorf("expected empty tasks list, got %d items", len(tasks))
	}
}

func TestCallTool_ComponentStatesList(t *testing.T) {
	srv := newTestServer(t)
	session := newTestSession(t, srv)

	result, err := session.CallTool(t.Context(), &gomcp.CallToolParams{
		Name: "component_states_list",
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}

	text, ok := result.Content[0].(*gomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var states []map[string]any
	if err := json.Unmarshal([]byte(text.Text), &states); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	// Should have entries for all registered components.
	if len(states) == 0 {
		t.Error("expected non-empty component states list")
	}
	// All should default to not_installed.
	for _, s := range states {
		if s["status"] != "not_installed" {
			t.Errorf("expected status not_installed, got %v for %v", s["status"], s["component"])
		}
	}
}

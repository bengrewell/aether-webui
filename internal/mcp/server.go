package mcp

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bengrewell/aether-webui/internal/provider/meta"
	"github.com/bengrewell/aether-webui/internal/provider/nodes"
	"github.com/bengrewell/aether-webui/internal/provider/onramp"
	"github.com/bengrewell/aether-webui/internal/provider/system"
	"github.com/bengrewell/aether-webui/internal/store"
)

// Config holds the dependencies required to create an MCP server.
type Config struct {
	Store   store.Client
	Nodes   *nodes.Nodes
	OnRamp  *onramp.OnRamp
	System  *system.System
	Meta    *meta.Meta
	Log     *slog.Logger
	Version string
}

// Server wraps an MCP server with registered tools backed by provider handlers.
type Server struct {
	srv    *mcp.Server
	store  store.Client
	nodes  *nodes.Nodes
	onramp *onramp.OnRamp
	system *system.System
	meta   *meta.Meta
	log    *slog.Logger
}

// New creates a configured MCP server with all tools registered.
func New(cfg Config) *Server {
	log := cfg.Log
	if log == nil {
		log = slog.Default()
	}

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "aether-webd",
		Version: cfg.Version,
	}, nil)

	s := &Server{
		srv:    srv,
		store:  cfg.Store,
		nodes:  cfg.Nodes,
		onramp: cfg.OnRamp,
		system: cfg.System,
		meta:   cfg.Meta,
		log:    log,
	}

	s.registerTools()
	return s
}

// registerTools registers all MCP tools across all domains.
func (s *Server) registerTools() {
	s.registerNodeTools()
	s.registerOnRampTools()
	s.registerTaskTools()
	s.registerSystemTools()
	s.registerMetaTools()
}

// RunStdio starts the MCP server on stdin/stdout. Blocks until ctx is done.
func (s *Server) RunStdio(ctx context.Context) error {
	s.log.Info("starting MCP server on stdio")
	return s.srv.Run(ctx, &mcp.StdioTransport{})
}

// HTTPHandler returns an http.Handler for StreamableHTTP transport.
func (s *Server) HTTPHandler() http.Handler {
	return mcp.NewStreamableHTTPHandler(
		func(_ *http.Request) *mcp.Server { return s.srv },
		nil,
	)
}

// MCPServer returns the underlying mcp.Server for testing and introspection.
func (s *Server) MCPServer() *mcp.Server {
	return s.srv
}

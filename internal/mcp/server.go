package mcp

import (
	"context"
	"sync"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/dcsg/archway/internal/provider"
)

type Server struct {
	mu             sync.RWMutex
	server         *gomcp.Server
	latestAnalysis *provider.AnalyzeResponse
}

func NewServer() *Server {
	s := &Server{}
	s.server = gomcp.NewServer(&gomcp.Implementation{
		Name:    "archway",
		Version: "v1",
	}, &gomcp.ServerOptions{
		Instructions: "Archway — Terraform for Code Architecture. Provides deterministic code analysis, architecture detection, and scaffolding tools.",
	})
	s.registerTools()
	s.registerResources()
	return s
}

func (s *Server) Serve(_ string) error {
	return s.server.Run(context.Background(), &gomcp.StdioTransport{})
}

func (s *Server) setLatestAnalysis(result *provider.AnalyzeResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latestAnalysis = result
}

func (s *Server) getLatestAnalysis() *provider.AnalyzeResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latestAnalysis
}

package mcp

import (
	"testing"
)

func TestNewServerCreates(t *testing.T) {
	s := NewServer()
	if s == nil {
		t.Fatal("expected non-nil server")
	}
	if s.server == nil {
		t.Fatal("expected non-nil underlying MCP server")
	}
}

func TestLatestAnalysisCache(t *testing.T) {
	s := NewServer()
	if result := s.getLatestAnalysis(); result != nil {
		t.Fatal("expected nil before any analysis")
	}
}

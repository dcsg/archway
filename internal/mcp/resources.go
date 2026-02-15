package mcp

import (
	"context"
	"encoding/json"
	"os"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/dcsg/archway/internal/config"
)

func (s *Server) registerResources() {
	s.server.AddResource(&gomcp.Resource{
		URI:         "archway://config",
		Name:        "Archway Config",
		Description: "Current archway.yaml project configuration.",
		MIMEType:    "application/json",
	}, func(_ context.Context, req *gomcp.ReadResourceRequest) (*gomcp.ReadResourceResult, error) {
		path, err := config.FindArchwayYAML(".")
		if err != nil {
			return &gomcp.ReadResourceResult{
				Contents: []*gomcp.ResourceContents{{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     `{"found": false}`,
				}},
			}, nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		result := map[string]any{
			"found":   true,
			"path":    path,
			"content": string(content),
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		return &gomcp.ReadResourceResult{
			Contents: []*gomcp.ResourceContents{{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(data),
			}},
		}, nil
	})

	s.server.AddResource(&gomcp.Resource{
		URI:         "archway://analysis",
		Name:        "Latest Analysis",
		Description: "Latest cached codebase analysis results.",
		MIMEType:    "application/json",
	}, func(_ context.Context, req *gomcp.ReadResourceRequest) (*gomcp.ReadResourceResult, error) {
		result := s.getLatestAnalysis()
		if result == nil {
			return &gomcp.ReadResourceResult{
				Contents: []*gomcp.ResourceContents{{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     `{"available": false}`,
				}},
			}, nil
		}
		data, err := json.MarshalIndent(map[string]any{
			"available": true,
			"result":    result,
		}, "", "  ")
		if err != nil {
			return nil, err
		}
		return &gomcp.ReadResourceResult{
			Contents: []*gomcp.ResourceContents{{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(data),
			}},
		}, nil
	})
}

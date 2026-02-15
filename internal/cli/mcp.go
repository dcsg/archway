package cli

import (
	"fmt"

	"github.com/dcsg/archway/internal/mcp"
	"github.com/spf13/cobra"
)

type mcpCommandOptions struct {
	Transport string
}

func newMCPCommand(_ *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server commands",
		Long:  "Run Archway as an MCP server to expose analysis and scaffolding tools.",
	}

	opts := &mcpCommandOptions{}
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start MCP server",
		Long:  "Start an MCP server over stdio transport.",
		Example: `  archway mcp serve
  archway mcp serve --transport stdio`,
		RunE: func(_ *cobra.Command, _ []string) error {
			if opts.Transport != "stdio" {
				return fmt.Errorf("unsupported transport %q (only stdio is supported)", opts.Transport)
			}
			return runMCPServe(opts)
		},
	}

	serveCmd.Flags().StringVar(&opts.Transport, "transport", "stdio", "Transport for MCP server (stdio)")
	cmd.AddCommand(serveCmd)

	return cmd
}

func runMCPServe(opts *mcpCommandOptions) error {
	server := mcp.NewServer()
	return server.Serve(opts.Transport)
}

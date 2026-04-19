package mcp

import (
	"context"
	"fmt"

	"github.com/LCSilent/helm-agent-eye/pkg/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initCRD() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("crd_analyze",
				mcp.WithDescription("Diagnose all CustomResourceDefinitions in the cluster, checking Established and NamesAccepted conditions"),
			),
			Handler: s.crdAnalyze,
		},
	}
}

func (s *Server) crdAnalyze(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	res, err := s.k8s.AnalyzeCRD(common.Request{
		Context: ctx,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to analyze CRDs: %v", err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

package mcp

import (
	"context"
	"fmt"

	"github.com/LCSilent/helm-agent-eye/pkg/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initDaemonSet() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("daemonset_analyze",
				mcp.WithDescription("Diagnose all daemonsets in a namespace, checking scheduled/ready pods and rollout status"),
				mcp.WithString("namespace",
					mcp.Required(),
					mcp.Description("the namespace to analyze daemonsets in"),
				),
			),
			Handler: s.daemonSetAnalyze,
		},
	}
}

func (s *Server) daemonSetAnalyze(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var ns string
	if v, ok := ctr.Params.Arguments["namespace"].(string); ok {
		ns = v
	}
	res, err := s.k8s.AnalyzeDaemonSet(common.Request{
		Context:   ctx,
		Namespace: ns,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to analyze daemonsets in namespace %s: %v", ns, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

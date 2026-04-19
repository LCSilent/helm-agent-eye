package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	helmclient "github.com/LCSilent/helm-agent-eye/pkg/helm"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initHelm() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("helm_install_and_diagnose",
				mcp.WithDescription("Install a Helm release and diagnose all Kubernetes resources in the target namespace"),
				mcp.WithString("release_name",
					mcp.Required(),
					mcp.Description("Name of the Helm release"),
				),
				mcp.WithString("chart",
					mcp.Required(),
					mcp.Description("Chart reference: local path, repo/chart-name, or OCI URL"),
				),
				mcp.WithString("namespace",
					mcp.Description("Target namespace (default: default)"),
				),
				mcp.WithString("values",
					mcp.Description("Values as a YAML string"),
				),
				mcp.WithString("set_values",
					mcp.Description("Comma-separated key=value pairs, e.g. image.tag=v1.0,replicas=2"),
				),
				mcp.WithString("version",
					mcp.Description("Chart version to install"),
				),
				mcp.WithBoolean("wait",
					mcp.Description("Wait for resources to be ready before diagnosing (default: true)"),
				),
				mcp.WithString("timeout",
					mcp.Description("Wait timeout duration, e.g. 5m, 10m (default: 5m)"),
				),
				mcp.WithBoolean("create_namespace",
					mcp.Description("Create the namespace if it does not exist (default: false)"),
				),
				mcp.WithBoolean("atomic",
					mcp.Description("If set, roll back on failure (default: false)"),
				),
			),
			Handler: s.helmInstallAndDiagnose,
		},
		{
			Tool: mcp.NewTool("helm_upgrade_and_diagnose",
				mcp.WithDescription("Upgrade an existing Helm release and diagnose all Kubernetes resources in the target namespace"),
				mcp.WithString("release_name",
					mcp.Required(),
					mcp.Description("Name of the Helm release to upgrade"),
				),
				mcp.WithString("chart",
					mcp.Required(),
					mcp.Description("Chart reference: local path, repo/chart-name, or OCI URL"),
				),
				mcp.WithString("namespace",
					mcp.Description("Target namespace (default: default)"),
				),
				mcp.WithString("values",
					mcp.Description("Values as a YAML string"),
				),
				mcp.WithString("set_values",
					mcp.Description("Comma-separated key=value pairs, e.g. image.tag=v1.1,replicas=3"),
				),
				mcp.WithString("version",
					mcp.Description("Chart version to upgrade to"),
				),
				mcp.WithBoolean("wait",
					mcp.Description("Wait for resources to be ready before diagnosing (default: true)"),
				),
				mcp.WithString("timeout",
					mcp.Description("Wait timeout duration, e.g. 5m, 10m (default: 5m)"),
				),
				mcp.WithBoolean("atomic",
					mcp.Description("If set, roll back on failure (default: false)"),
				),
				mcp.WithBoolean("reuse_values",
					mcp.Description("Reuse the last release's values and merge with any overrides (default: false)"),
				),
				mcp.WithBoolean("force",
					mcp.Description("Force resource updates through a replacement strategy (default: false)"),
				),
				mcp.WithBoolean("cleanup_on_fail",
					mcp.Description("Allow deletion of new resources created in this upgrade when upgrade fails (default: false)"),
				),
			),
			Handler: s.helmUpgradeAndDiagnose,
		},
	}
}

func (s *Server) helmInstallAndDiagnose(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	opts := helmclient.InstallOptions{
		ReleaseName: ctr.Params.Arguments["release_name"].(string),
		Chart:       ctr.Params.Arguments["chart"].(string),
		Wait:        true,
		Timeout:     5 * time.Minute,
	}

	if ns, ok := ctr.Params.Arguments["namespace"].(string); ok {
		opts.Namespace = ns
	}
	if v, ok := ctr.Params.Arguments["values"].(string); ok {
		opts.Values = v
	}
	if sv, ok := ctr.Params.Arguments["set_values"].(string); ok {
		opts.SetValues = sv
	}
	if ver, ok := ctr.Params.Arguments["version"].(string); ok {
		opts.Version = ver
	}
	if wait, ok := ctr.Params.Arguments["wait"].(bool); ok {
		opts.Wait = wait
	}
	if t, ok := ctr.Params.Arguments["timeout"].(string); ok && t != "" {
		if d, err := time.ParseDuration(t); err == nil {
			opts.Timeout = d
		}
	}
	if cn, ok := ctr.Params.Arguments["create_namespace"].(bool); ok {
		opts.CreateNamespace = cn
	}
	if atomic, ok := ctr.Params.Arguments["atomic"].(bool); ok {
		opts.Atomic = atomic
	}

	result, err := s.helmClient.InstallAndDiagnose(ctx, opts)
	if err != nil && result == nil {
		return mcp.NewToolResultError(fmt.Sprintf("helm install failed: %v", err)), nil
	}

	data, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", jsonErr)), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) helmUpgradeAndDiagnose(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	opts := helmclient.UpgradeOptions{
		ReleaseName: ctr.Params.Arguments["release_name"].(string),
		Chart:       ctr.Params.Arguments["chart"].(string),
		Wait:        true,
		Timeout:     5 * time.Minute,
	}

	if ns, ok := ctr.Params.Arguments["namespace"].(string); ok {
		opts.Namespace = ns
	}
	if v, ok := ctr.Params.Arguments["values"].(string); ok {
		opts.Values = v
	}
	if sv, ok := ctr.Params.Arguments["set_values"].(string); ok {
		opts.SetValues = sv
	}
	if ver, ok := ctr.Params.Arguments["version"].(string); ok {
		opts.Version = ver
	}
	if wait, ok := ctr.Params.Arguments["wait"].(bool); ok {
		opts.Wait = wait
	}
	if t, ok := ctr.Params.Arguments["timeout"].(string); ok && t != "" {
		if d, err := time.ParseDuration(t); err == nil {
			opts.Timeout = d
		}
	}
	if atomic, ok := ctr.Params.Arguments["atomic"].(bool); ok {
		opts.Atomic = atomic
	}
	if rv, ok := ctr.Params.Arguments["reuse_values"].(bool); ok {
		opts.ReuseValues = rv
	}
	if force, ok := ctr.Params.Arguments["force"].(bool); ok {
		opts.Force = force
	}
	if cof, ok := ctr.Params.Arguments["cleanup_on_fail"].(bool); ok {
		opts.CleanupOnFail = cof
	}

	result, err := s.helmClient.UpgradeAndDiagnose(ctx, opts)
	if err != nil && result == nil {
		return mcp.NewToolResultError(fmt.Sprintf("helm upgrade failed: %v", err)), nil
	}

	data, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", jsonErr)), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

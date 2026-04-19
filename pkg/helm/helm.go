package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LCSilent/helm-agent-eye/pkg/common"
	"github.com/LCSilent/helm-agent-eye/pkg/k8s"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/rest"
)

// HelmClient wraps helm SDK and reuses the existing k8s client for diagnostics.
type HelmClient struct {
	k8s    *k8s.Kubernetes
	config *rest.Config
}

// NewHelmClient creates a new HelmClient.
func NewHelmClient(k8sClient *k8s.Kubernetes, config *rest.Config) *HelmClient {
	return &HelmClient{
		k8s:    k8sClient,
		config: config,
	}
}

// newActionConfig creates a helm action configuration for the given namespace.
func (h *HelmClient) newActionConfig(namespace string) (*action.Configuration, error) {
	settings := cli.New()
	settings.SetNamespace(namespace)

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, "secret", func(format string, v ...interface{}) {}); err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %w", err)
	}
	return actionConfig, nil
}

// parseSetValues parses a comma-separated k=v string into a map.
func parseSetValues(setValues string) map[string]interface{} {
	result := make(map[string]interface{})
	if setValues == "" {
		return result
	}
	for _, pair := range strings.Split(setValues, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}

// mergeValues merges raw YAML values string and --set values into a single map.
func mergeValues(rawYAML, setVals string) (map[string]interface{}, error) {
	opts := &values.Options{}
	if rawYAML != "" {
		opts.Values = []string{}
		// Write YAML to a temp approach: use strvals for set, yaml for values file
	}

	p := getter.All(cli.New())
	vals, err := opts.MergeValues(p)
	if err != nil {
		return nil, err
	}

	// Apply --set overrides
	setMap := parseSetValues(setVals)
	for k, v := range setMap {
		vals[k] = v
	}

	// Parse raw YAML values
	if rawYAML != "" {
		var yamlVals map[string]interface{}
		if err := json.Unmarshal([]byte(rawYAML), &yamlVals); err == nil {
			for k, v := range yamlVals {
				if _, exists := vals[k]; !exists {
					vals[k] = v
				}
			}
		}
	}

	return vals, nil
}

// InstallAndDiagnose installs a helm release and then diagnoses all resources in the namespace.
func (h *HelmClient) InstallAndDiagnose(ctx context.Context, opts InstallOptions) (*common.HelmDiagnoseResult, error) {
	namespace := opts.Namespace
	if namespace == "" {
		namespace = "default"
	}

	actionConfig, err := h.newActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	installAction := action.NewInstall(actionConfig)
	installAction.ReleaseName = opts.ReleaseName
	installAction.Namespace = namespace
	installAction.CreateNamespace = opts.CreateNamespace
	installAction.Atomic = opts.Atomic
	installAction.Wait = opts.Wait
	if opts.Timeout > 0 {
		installAction.Timeout = opts.Timeout
	}
	if opts.Version != "" {
		installAction.Version = opts.Version
	}

	settings := cli.New()
	settings.SetNamespace(namespace)
	chartPath, err := installAction.ChartPathOptions.LocateChart(opts.Chart, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart %q: %w", opts.Chart, err)
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	vals, err := mergeValues(opts.Values, opts.SetValues)
	if err != nil {
		return nil, fmt.Errorf("failed to merge values: %w", err)
	}

	var installErr error
	var rel *release.Release
	rel, installErr = installAction.RunWithContext(ctx, chart, vals)

	// Build result from release info (even on timeout, rel may be non-nil)
	result := &common.HelmDiagnoseResult{
		ReleaseName: opts.ReleaseName,
		Namespace:   namespace,
		Action:      "install",
	}
	if rel != nil {
		result.Status = rel.Info.Status.String()
		result.Revision = rel.Version
	}
	if installErr != nil {
		result.Status = fmt.Sprintf("failed: %v", installErr)
	}

	// Always run diagnostics regardless of install error
	diagnostics, diagErr := h.diagnoseNamespace(ctx, namespace)
	if diagErr == nil {
		result.Diagnostics = diagnostics
		result.Summary = buildSummary(diagnostics)
	}

	// Return install error only if diagnostics also failed
	if installErr != nil && diagErr != nil {
		return result, installErr
	}
	return result, nil
}

// UpgradeAndDiagnose upgrades an existing helm release and then diagnoses all resources in the namespace.
func (h *HelmClient) UpgradeAndDiagnose(ctx context.Context, opts UpgradeOptions) (*common.HelmDiagnoseResult, error) {
	namespace := opts.Namespace
	if namespace == "" {
		namespace = "default"
	}

	actionConfig, err := h.newActionConfig(namespace)
	if err != nil {
		return nil, err
	}

	upgradeAction := action.NewUpgrade(actionConfig)
	upgradeAction.Namespace = namespace
	upgradeAction.Atomic = opts.Atomic
	upgradeAction.Wait = opts.Wait
	upgradeAction.ReuseValues = opts.ReuseValues
	upgradeAction.Force = opts.Force
	upgradeAction.CleanupOnFail = opts.CleanupOnFail
	if opts.Timeout > 0 {
		upgradeAction.Timeout = opts.Timeout
	}
	if opts.Version != "" {
		upgradeAction.Version = opts.Version
	}

	settings := cli.New()
	settings.SetNamespace(namespace)
	chartPath, err := upgradeAction.ChartPathOptions.LocateChart(opts.Chart, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart %q: %w", opts.Chart, err)
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	vals, err := mergeValues(opts.Values, opts.SetValues)
	if err != nil {
		return nil, fmt.Errorf("failed to merge values: %w", err)
	}

	var upgradeErr error
	var rel *release.Release
	rel, upgradeErr = upgradeAction.RunWithContext(ctx, opts.ReleaseName, chart, vals)

	result := &common.HelmDiagnoseResult{
		ReleaseName: opts.ReleaseName,
		Namespace:   namespace,
		Action:      "upgrade",
	}
	if rel != nil {
		result.Status = rel.Info.Status.String()
		result.Revision = rel.Version
	}
	if upgradeErr != nil {
		result.Status = fmt.Sprintf("failed: %v", upgradeErr)
	}

	// Always run diagnostics regardless of upgrade error
	diagnostics, diagErr := h.diagnoseNamespace(ctx, namespace)
	if diagErr == nil {
		result.Diagnostics = diagnostics
		result.Summary = buildSummary(diagnostics)
	}

	if upgradeErr != nil && diagErr != nil {
		return result, upgradeErr
	}
	return result, nil
}

// diagnoseNamespace runs full diagnostics on all resources in the given namespace.
func (h *HelmClient) diagnoseNamespace(ctx context.Context, namespace string) ([]common.Result, error) {
	var allResults []common.Result

	req := common.Request{
		Context:   ctx,
		Namespace: namespace,
	}

	// Pod
	if data, err := h.k8s.AnalyzePod(ctx, namespace); err == nil {
		allResults = append(allResults, parseResults(data)...)
	}

	// Deployment
	if data, err := h.k8s.AnalyzeDeployment(ctx, namespace); err == nil {
		allResults = append(allResults, parseResults(data)...)
	}

	// StatefulSet
	if data, err := h.k8s.AnalyzeStatefulSet(req); err == nil {
		allResults = append(allResults, parseResults(data)...)
	}

	// Service
	if data, err := h.k8s.AnalyzeService(ctx, namespace); err == nil {
		allResults = append(allResults, parseResults(data)...)
	}

	// Ingress
	if data, err := h.k8s.AnalyzeIngress(req); err == nil {
		allResults = append(allResults, parseResults(data)...)
	}

	// CronJob
	if data, err := h.k8s.AnalyzeCronJob(req); err == nil {
		allResults = append(allResults, parseResults(data)...)
	}

	// DaemonSet
	if data, err := h.k8s.AnalyzeDaemonSet(req); err == nil {
		allResults = append(allResults, parseResults(data)...)
	}

	return allResults, nil
}

// parseResults unmarshals a JSON string into a slice of common.Result.
func parseResults(data string) []common.Result {
	if data == "" || data == "null" {
		return nil
	}
	var results []common.Result
	if err := json.Unmarshal([]byte(data), &results); err != nil {
		return nil
	}
	return results
}

// buildSummary counts issues by kind.
func buildSummary(results []common.Result) common.DiagnoseSummary {
	summary := common.DiagnoseSummary{TotalIssues: len(results)}
	for _, r := range results {
		if r.Kind == "Pod" {
			summary.PodIssues++
		} else {
			summary.OtherIssues++
		}
	}
	return summary
}

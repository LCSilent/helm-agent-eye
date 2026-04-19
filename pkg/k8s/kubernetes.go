// Package k8s provides a unified facade over all Kubernetes resource analyzers.
// External callers use this package exclusively; the sub-packages (base, core, apps, …)
// are internal implementation details.
package k8s

import (
	"context"

	"github.com/LCSilent/helm-agent-eye/pkg/common"
	"github.com/LCSilent/helm-agent-eye/pkg/k8s/admissionregistration"
	"github.com/LCSilent/helm-agent-eye/pkg/k8s/apiextensions"
	"github.com/LCSilent/helm-agent-eye/pkg/k8s/apps"
	"github.com/LCSilent/helm-agent-eye/pkg/k8s/base"
	"github.com/LCSilent/helm-agent-eye/pkg/k8s/batch"
	"github.com/LCSilent/helm-agent-eye/pkg/k8s/core"
	"github.com/LCSilent/helm-agent-eye/pkg/k8s/networking"
	"k8s.io/client-go/rest"
)

// Kubernetes is the top-level client that aggregates all sub-analyzers.
// It embeds *base.Kubernetes so that resource-level methods (ResourceList,
// ResourceGet, …) are promoted directly.
type Kubernetes struct {
	*base.Kubernetes

	coreAnalyzer          *core.Analyzer
	appsAnalyzer          *apps.Analyzer
	batchAnalyzer         *batch.Analyzer
	networkingAnalyzer    *networking.Analyzer
	admissionAnalyzer     *admissionregistration.Analyzer
	apiextensionsAnalyzer *apiextensions.Analyzer
}

// NewKubernetes creates a fully-initialised Kubernetes facade.
func NewKubernetes() (*Kubernetes, error) {
	k, err := base.NewKubernetes()
	if err != nil {
		return nil, err
	}
	return &Kubernetes{
		Kubernetes:            k,
		coreAnalyzer:          core.NewAnalyzer(k),
		appsAnalyzer:          apps.NewAnalyzer(k),
		batchAnalyzer:         batch.NewAnalyzer(k),
		networkingAnalyzer:    networking.NewAnalyzer(k),
		admissionAnalyzer:     admissionregistration.NewAnalyzer(k),
		apiextensionsAnalyzer: apiextensions.NewAnalyzer(k),
	}, nil
}

// RestConfig returns the underlying *rest.Config.
func (k *Kubernetes) RestConfig() *rest.Config {
	return k.Kubernetes.RestConfig()
}

// ── core/v1 ──────────────────────────────────────────────────────────────────

func (k *Kubernetes) PodLogs(ctx context.Context, namespace, name string) (string, error) {
	return k.coreAnalyzer.PodLogs(ctx, namespace, name)
}

func (k *Kubernetes) PodExec(ctx context.Context, namespace, name, command string) (string, error) {
	return k.coreAnalyzer.PodExec(ctx, namespace, name, command)
}

func (k *Kubernetes) AnalyzePod(ctx context.Context, namespace string) (string, error) {
	return k.coreAnalyzer.AnalyzePod(ctx, namespace)
}

func (k *Kubernetes) AnalyzeNode(ctx context.Context, name string) (string, error) {
	return k.coreAnalyzer.AnalyzeNode(ctx, name)
}

func (k *Kubernetes) AnalyzeService(ctx context.Context, namespace string) (string, error) {
	return k.coreAnalyzer.AnalyzeService(ctx, namespace)
}

// ── apps/v1 ──────────────────────────────────────────────────────────────────

func (k *Kubernetes) DeploymentScale(ctx context.Context, namespace, name string, replicas int32) (string, error) {
	return k.appsAnalyzer.DeploymentScale(ctx, namespace, name, replicas)
}

func (k *Kubernetes) AnalyzeDeployment(ctx context.Context, namespace string) (string, error) {
	return k.appsAnalyzer.AnalyzeDeployment(ctx, namespace)
}

func (k *Kubernetes) AnalyzeStatefulSet(r common.Request) (string, error) {
	return k.appsAnalyzer.AnalyzeStatefulSet(r)
}

func (k *Kubernetes) AnalyzeDaemonSet(r common.Request) (string, error) {
	return k.appsAnalyzer.AnalyzeDaemonSet(r)
}

// ── batch/v1 ─────────────────────────────────────────────────────────────────

func (k *Kubernetes) AnalyzeCronJob(r common.Request) (string, error) {
	return k.batchAnalyzer.AnalyzeCronJob(r)
}

// ── networking.k8s.io ────────────────────────────────────────────────────────

func (k *Kubernetes) AnalyzeIngress(r common.Request) (string, error) {
	return k.networkingAnalyzer.AnalyzeIngress(r)
}

func (k *Kubernetes) AnalyzeNetworkPolicy(r common.Request) (string, error) {
	return k.networkingAnalyzer.AnalyzeNetworkPolicy(r)
}

// ── admissionregistration.k8s.io ─────────────────────────────────────────────

func (k *Kubernetes) AnalyzeValidatingWebhook(r common.Request) (string, error) {
	return k.admissionAnalyzer.AnalyzeValidatingWebhook(r)
}

func (k *Kubernetes) AnalyzeMutatingWebhook(r common.Request) (string, error) {
	return k.admissionAnalyzer.AnalyzeMutatingWebhook(r)
}

// ── apiextensions.k8s.io ─────────────────────────────────────────────────────

func (k *Kubernetes) AnalyzeCRD(r common.Request) (string, error) {
	return k.apiextensionsAnalyzer.AnalyzeCRD(r)
}

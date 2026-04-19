package core

import (
	"github.com/LCSilent/helm-agent-eye/pkg/k8s/base"
)

// Analyzer provides analysis methods for core/v1 resources.
type Analyzer struct {
	k8s *base.Kubernetes
}

// NewAnalyzer creates a new core Analyzer.
func NewAnalyzer(k *base.Kubernetes) *Analyzer {
	return &Analyzer{k8s: k}
}

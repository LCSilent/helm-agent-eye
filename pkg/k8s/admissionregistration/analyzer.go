package admissionregistration

import (
	"github.com/LCSilent/helm-agent-eye/pkg/k8s/base"
)

// Analyzer provides analysis methods for admissionregistration.k8s.io resources.
type Analyzer struct {
	k8s *base.Kubernetes
}

// NewAnalyzer creates a new admissionregistration Analyzer.
func NewAnalyzer(k *base.Kubernetes) *Analyzer {
	return &Analyzer{k8s: k}
}

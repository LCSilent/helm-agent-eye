package helm

import "time"

// InstallOptions contains options for helm install operation.
type InstallOptions struct {
	ReleaseName     string
	Chart           string
	Namespace       string
	Values          string // raw YAML string
	SetValues       string // comma-separated k=v pairs
	Version         string
	Wait            bool
	Timeout         time.Duration
	CreateNamespace bool
	Atomic          bool
}

// UpgradeOptions contains options for helm upgrade operation.
type UpgradeOptions struct {
	ReleaseName   string
	Chart         string
	Namespace     string
	Values        string // raw YAML string
	SetValues     string // comma-separated k=v pairs
	Version       string
	Wait          bool
	Timeout       time.Duration
	Atomic        bool
	ReuseValues   bool
	Force         bool
	CleanupOnFail bool
}

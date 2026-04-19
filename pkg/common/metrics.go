package common

type WorkloadFormattedMetrics struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	CPUUsage    string `json:"cpu"`
	MemoryUsage string `json:"memory"`
}

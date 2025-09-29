package config

// Config holds the agent configuration
type Config struct {
	// Output directory for event JSON files
	OutputDir string `json:"output_dir"`

	// Sampling duration in seconds (how long to collect)
	SamplingOn int `json:"sampling_on"`

	// Sampling off duration in seconds (how long to wait between collections)
	SamplingOff int `json:"sampling_off"`

	// Optional HTTP endpoint to send events to
	HTTPEndpoint string `json:"http_endpoint,omitempty"`

	// Verbose logging
	Verbose bool `json:"verbose"`

	// Enable specific monitors
	MonitorProcesses bool `json:"monitor_processes"`
	MonitorSyscalls  bool `json:"monitor_syscalls"`
	MonitorNetwork   bool `json:"monitor_network"`

	// Container runtime socket (for metadata enrichment)
	ContainerRuntime string `json:"container_runtime"` // "docker", "containerd", etc.
}

// Default returns a default configuration
func Default() *Config {
	return &Config{
		OutputDir:        "/var/log/sds-observer",
		SamplingOn:       60,
		SamplingOff:      240,
		Verbose:          false,
		MonitorProcesses: true,
		MonitorSyscalls:  true,
		MonitorNetwork:   true,
		ContainerRuntime: "docker",
	}
}
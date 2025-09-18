package context

// SystemContext holds system-level information
type SystemContext struct {
	OperatingSystem string // Operating system (runtime.GOOS)
	Architecture    string // CPU architecture (runtime.GOARCH)
	MachineID       string // Unique machine identifier
	Hostname        string // System hostname
	Username        string // Current user's username
	MaxConcurrency  int    // Maximum concurrency level
}

// NewSystemContext creates a new system context
func NewSystemContext(os, arch, machineID, hostname, username string, maxConcurrency int) *SystemContext {
	return &SystemContext{
		OperatingSystem: os,
		Architecture:    arch,
		MachineID:       machineID,
		Hostname:        hostname,
		Username:        username,
		MaxConcurrency:  maxConcurrency,
	}
}

// GetOS returns the operating system
func (s *SystemContext) GetOS() string {
	return s.OperatingSystem
}

// GetArchitecture returns the CPU architecture
func (s *SystemContext) GetArchitecture() string {
	return s.Architecture
}

// GetMachineID returns the machine ID
func (s *SystemContext) GetMachineID() string {
	return s.MachineID
}

// GetHostname returns the hostname
func (s *SystemContext) GetHostname() string {
	return s.Hostname
}

// GetUsername returns the username
func (s *SystemContext) GetUsername() string {
	return s.Username
}

// GetMaxConcurrency returns the maximum concurrency level
func (s *SystemContext) GetMaxConcurrency() int {
	return s.MaxConcurrency
}
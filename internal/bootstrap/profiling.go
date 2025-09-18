package bootstrap

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

// ProfilingManager handles CPU and memory profiling
type ProfilingManager struct {
	cpuProfileFile *os.File
	memProfileFile string
	enabled        bool
}

// NewProfilingManager creates a new profiling manager
func NewProfilingManager(cpuProfile, memProfile string) *ProfilingManager {
	return &ProfilingManager{
		memProfileFile: memProfile,
		enabled:        cpuProfile != "" || memProfile != "",
	}
}

// StartCPUProfiling starts CPU profiling if configured
func (p *ProfilingManager) StartCPUProfiling(cpuProfile string) error {
	if cpuProfile == "" {
		return nil
	}

	f, err := os.Create(cpuProfile)
	if err != nil {
		return fmt.Errorf("could not create CPU profile: %w", err)
	}

	p.cpuProfileFile = f

	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return fmt.Errorf("could not start CPU profile: %w", err)
	}

	return nil
}

// StopCPUProfiling stops CPU profiling if it was started
func (p *ProfilingManager) StopCPUProfiling() {
	if p.cpuProfileFile != nil {
		pprof.StopCPUProfile()
		p.cpuProfileFile.Close()
		p.cpuProfileFile = nil
	}
}

// WriteMemProfile writes memory profile if configured
func (p *ProfilingManager) WriteMemProfile() error {
	if p.memProfileFile == "" {
		return nil
	}

	f, err := os.Create(p.memProfileFile)
	if err != nil {
		return fmt.Errorf("could not create memory profile: %w", err)
	}
	defer f.Close()

	runtime.GC() // get up-to-date statistics

	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("could not write memory profile: %w", err)
	}

	return nil
}

// IsEnabled returns true if any profiling is enabled
func (p *ProfilingManager) IsEnabled() bool {
	return p.enabled
}

// Cleanup ensures all profiling is properly stopped
func (p *ProfilingManager) Cleanup() error {
	p.StopCPUProfiling()

	if err := p.WriteMemProfile(); err != nil {
		return err
	}

	return nil
}
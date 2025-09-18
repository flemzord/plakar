package context

// ExecutionContext holds runtime execution information
type ExecutionContext struct {
	CWD         string // Current working directory
	ProcessID   int    // Process ID
	CommandLine string // Full command line used to invoke the program
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(cwd string, pid int, cmdLine string) *ExecutionContext {
	return &ExecutionContext{
		CWD:         cwd,
		ProcessID:   pid,
		CommandLine: cmdLine,
	}
}

// GetCWD returns the current working directory
func (e *ExecutionContext) GetCWD() string {
	return e.CWD
}

// GetProcessID returns the process ID
func (e *ExecutionContext) GetProcessID() int {
	return e.ProcessID
}

// GetCommandLine returns the full command line
func (e *ExecutionContext) GetCommandLine() string {
	return e.CommandLine
}
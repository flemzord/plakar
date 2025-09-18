package bootstrap

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/PlakarKorp/plakar/agent"
	"github.com/PlakarKorp/plakar/subcommands"
	"github.com/PlakarKorp/plakar/task"
	"github.com/PlakarKorp/plakar/utils"
	"errors"
)

// Pipeline orchestrates the application bootstrap process
type Pipeline struct {
	config     *ConfigContext
	profiling  *ProfilingManager
	security   *SecurityManager
	repository *RepositoryManager
	signals    *SignalHandler

	cmd        subcommands.Subcommand
	cmdName    []string
	cmdArgs    []string

	stdout     io.Writer
	stderr     io.Writer

	startTime  time.Time
}

// NewPipeline creates a new bootstrap pipeline
func NewPipeline(stdout, stderr io.Writer) *Pipeline {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	return &Pipeline{
		stdout:    stdout,
		stderr:    stderr,
		startTime: time.Now(),
	}
}

// WithConfig initializes configuration
func (p *Pipeline) WithConfig() (*Pipeline, error) {
	opts, err := NewOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to create options: %w", err)
	}

	if err := opts.ParseFlags(); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	config, err := InitializeContext(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize context: %w", err)
	}

	p.config = config
	return p, nil
}

// WithProfiling sets up profiling if configured
func (p *Pipeline) WithProfiling() (*Pipeline, error) {
	if p.config == nil {
		return nil, fmt.Errorf("config must be initialized before profiling")
	}

	p.profiling = NewProfilingManager(
		p.config.Options.CPUProfile,
		p.config.Options.MemProfile,
	)

	if err := p.profiling.StartCPUProfiling(p.config.Options.CPUProfile); err != nil {
		return nil, err
	}

	return p, nil
}

// WithSecurity handles security checks
func (p *Pipeline) WithSecurity() (*Pipeline, error) {
	if p.config == nil {
		return nil, fmt.Errorf("config must be initialized before security")
	}

	p.security = NewSecurityManager(p.config.CacheDir, p.stdout, p.stderr)

	// Handle security flags - returns true if we should exit
	if p.security.HandleSecurityFlags(
		p.config.AppContext,
		p.config.Options.EnableSecurityCheck,
		p.config.Options.DisableSecurityCheck,
	) {
		return nil, fmt.Errorf("security flag handled, exiting")
	}

	// Check for updates
	p.security.CheckForUpdates(p.config.AppContext)

	return p, nil
}

// WithRepository initializes repository if needed
func (p *Pipeline) WithRepository() (*Pipeline, error) {
	if p.config == nil {
		return nil, fmt.Errorf("config must be initialized before repository")
	}

	// Parse repository path
	if err := p.config.ParseRepository(); err != nil {
		return nil, err
	}

	// Get passphrase from environment
	passphrase, err := GetPassphraseFromEnv(p.config.AppContext, p.config.StoreConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get passphrase: %w", err)
	}
	if passphrase != "" {
		p.config.KeyFromFile = passphrase
	}

	// Lookup command
	cmd, name, args := subcommands.Lookup(p.config.Args)
	if cmd == nil {
		return nil, fmt.Errorf("command not found: %s", p.config.Args[0])
	}

	p.cmd = cmd
	p.cmdName = name
	p.cmdArgs = args

	// Load plugins
	if err := p.config.GetPlugins().LoadPlugins(p.config.GetInner()); err != nil {
		return nil, fmt.Errorf("failed to load plugins: %w", err)
	}

	// Initialize repository
	p.repository = NewRepositoryManager(p.config)
	if err := p.repository.InitializeRepository(cmd, p.config.Options.Agentless); err != nil {
		return nil, err
	}

	return p, nil
}

// WithSignals sets up signal handling
func (p *Pipeline) WithSignals() (*Pipeline, error) {
	p.signals = NewSignalHandler(p.stderr)
	p.signals.Start(func() {
		if p.config != nil {
			p.config.Cancel()
		}
	})

	return p, nil
}

// Execute runs the command
func (p *Pipeline) Execute() int {
	// Ensure cleanup happens
	defer p.cleanup()

	// Parse command arguments
	if err := p.cmd.Parse(p.config.AppContext, p.cmdArgs); err != nil {
		fmt.Fprintf(p.stderr, "Error: %s\n", err)
		return 1
	}

	// Set command metadata
	p.cmd.SetCWD(p.config.CWD)
	p.cmd.SetCommandLine(p.config.CommandLine)

	// Execute command
	var status int
	var err error

	runWithoutAgent := p.config.Options.Agentless || p.cmd.GetFlags()&subcommands.AgentSupport == 0
	if runWithoutAgent {
		status, err = task.RunCommand(
			p.config.AppContext,
			p.cmd,
			p.repository.GetRepository(),
			"@agentless",
		)
	} else {
		status, err = agent.ExecuteRPC(
			p.config.AppContext,
			p.cmdName,
			p.cmd,
			p.config.StoreConfig,
		)
	}

	if err != nil {
		fmt.Fprintf(p.stderr, "Error: %s\n", utils.SanitizeText(err.Error()))
		if errors.Is(err, agent.ErrWrongVersion) {
			fmt.Fprintln(p.stderr, "To stop the current agent, run:")
			fmt.Fprintln(p.stderr, "\t$ plakar agent stop")
		}
		if status == 0 {
			status = 1
		}
	}

	// Display execution time if requested
	if p.config.Options.Time {
		fmt.Fprintf(p.stdout, "time: %v\n", time.Since(p.startTime))
	}

	return status
}

// cleanup ensures all resources are properly released
func (p *Pipeline) cleanup() {
	// Stop signal handler
	if p.signals != nil {
		p.signals.Cleanup()
	}

	// Close repository
	if p.repository != nil {
		if err := p.repository.Close(); err != nil {
			// Log but don't fail
			if p.config != nil && p.config.GetLogger() != nil {
				p.config.GetLogger().Warn("Repository cleanup error: %s", err)
			}
		}
	}

	// Stop profiling
	if p.profiling != nil {
		if err := p.profiling.Cleanup(); err != nil {
			fmt.Fprintf(p.stderr, "Profiling cleanup error: %s\n", err)
		}
	}

	// Close context resources
	if p.config != nil {
		p.config.Close()
	}
}

// ListCommands displays available commands (helper function)
func ListCommands(out io.Writer, prefix string) {
	var last string
	var subs []string

	flush := func() {
		pre, post := " ", ""
		if len(subs) > 1 && subs[0] == "" {
			pre, post = " [", "]"
			subs = subs[1:]
		}
		subcmds := ""
		for i, s := range subs {
			if i > 0 {
				subcmds += " | "
			}
			subcmds += s
		}
		fmt.Fprint(out, prefix, last, pre, subcmds, post, "\n")
	}

	all := subcommands.List()
	for _, cmd := range all {
		if len(cmd) == 0 || cmd[0] == "diag" {
			continue
		}

		if last == "" {
			goto next
		}

		if last == cmd[0] {
			if len(subs) > 0 && subs[len(subs)-1] != cmd[1] {
				subs = append(subs, cmd[1])
			}
			continue
		}

		flush()

	next:
		subs = subs[:0]
		last = cmd[0]
		if len(cmd) > 1 {
			subs = append(subs, cmd[1])
		} else {
			subs = append(subs, "")
		}
	}
	flush()
}
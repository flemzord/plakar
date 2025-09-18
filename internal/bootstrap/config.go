package bootstrap

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/PlakarKorp/plakar/appcontext"
	"github.com/PlakarKorp/plakar/cookies"
	"github.com/PlakarKorp/plakar/plugins"
	"github.com/PlakarKorp/plakar/utils"
	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
	"github.com/PlakarKorp/kloset/caching"
	"github.com/PlakarKorp/kloset/logging"
)

// Options represents all command line options
type Options struct {
	ConfigDir             string
	CPUCount             int
	CPUProfile           string
	MemProfile           string
	Time                 bool
	Trace                string
	Quiet                bool
	KeyFile              string
	Agentless            bool
	EnableSecurityCheck  bool
	DisableSecurityCheck bool
}

// ConfigContext holds all configuration related data
type ConfigContext struct {
	*appcontext.AppContext
	Options       *Options
	RepositoryPath string
	StoreConfig   map[string]string
	Args          []string
	IsAt          bool
}

// NewOptions creates Options with default values
func NewOptions() (*Options, error) {
	cpuDefault := runtime.GOMAXPROCS(0)
	if cpuDefault != 1 {
		cpuDefault = cpuDefault - 1
	}

	configDefault, err := utils.GetConfigDir("plakar")
	if err != nil {
		return nil, fmt.Errorf("could not get config directory: %w", err)
	}

	return &Options{
		ConfigDir: configDefault,
		CPUCount:  cpuDefault,
	}, nil
}

// ParseFlags parses command line flags
func (o *Options) ParseFlags() error {
	flag.StringVar(&o.ConfigDir, "config", o.ConfigDir, "configuration directory")
	flag.IntVar(&o.CPUCount, "cpu", o.CPUCount, "limit the number of usable cores")
	flag.StringVar(&o.CPUProfile, "profile-cpu", "", "profile CPU usage")
	flag.StringVar(&o.MemProfile, "profile-mem", "", "profile MEM usage")
	flag.BoolVar(&o.Time, "time", false, "display command execution time")
	flag.StringVar(&o.Trace, "trace", "", "display trace logs, comma-separated (all, trace, repository, snapshot, server)")
	flag.BoolVar(&o.Quiet, "quiet", false, "no output except errors")
	flag.StringVar(&o.KeyFile, "keyfile", "", "use passphrase from key file when prompted")
	flag.BoolVar(&o.Agentless, "no-agent", false, "run without agent")
	flag.BoolVar(&o.EnableSecurityCheck, "enable-security-check", false, "enable update check")
	flag.BoolVar(&o.DisableSecurityCheck, "disable-security-check", false, "disable update check")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTIONS] [at REPOSITORY] COMMAND [COMMAND_OPTIONS]...\n", flag.CommandLine.Name())
		fmt.Fprintf(flag.CommandLine.Output(), "\nBy default, the repository is $PLAKAR_REPOSITORY or $HOME/.plakar.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\nOPTIONS:\n")
		flag.PrintDefaults()
		// Note: listCmds function will need to be imported or passed as parameter
	}

	flag.Parse()

	// Validate CPU count
	if o.CPUCount <= 0 {
		return fmt.Errorf("invalid -cpu value %d", o.CPUCount)
	}
	if o.CPUCount > runtime.NumCPU() {
		return fmt.Errorf("can't use more cores than available: %d", runtime.NumCPU())
	}

	return nil
}

// InitializeContext creates and configures the AppContext
func InitializeContext(opts *Options) (*ConfigContext, error) {
	ctx := appcontext.NewAppContext()

	ctx.ConfigDir = opts.ConfigDir
	if err := ctx.ReloadConfig(); err != nil {
		return nil, fmt.Errorf("could not load configuration: %w", err)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Get user info
	currentUser, err := user.Current()
	if err != nil {
		return nil, errors.New("cannot determine current user")
	}

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	// Get machine ID
	machineID, err := machineid.ID()
	if err != nil {
		machineID = uuid.NewSHA1(uuid.Nil, []byte(hostname)).String()
	}
	machineID = strings.ToLower(machineID)

	// Set context properties
	ctx.Client = "plakar/" + utils.GetVersion()
	ctx.CWD = cwd
	ctx.OperatingSystem = runtime.GOOS
	ctx.Architecture = runtime.GOARCH
	ctx.Username = currentUser.Username
	ctx.Hostname = hostname
	ctx.CommandLine = strings.Join(os.Args, " ")
	ctx.MachineID = machineID
	ctx.ProcessID = os.Getpid()
	ctx.MaxConcurrency = opts.CPUCount*2 + 1

	// Handle agentless mode
	_, envAgentLess := os.LookupEnv("PLAKAR_AGENTLESS")
	if envAgentLess || runtime.GOOS == "windows" {
		opts.Agentless = true
	}

	// Initialize cache
	cacheSubDir := "plakar"
	if opts.Agentless {
		cacheSubDir = "plakar-agentless"
	}

	cacheDir, err := utils.GetCacheDir(cacheSubDir)
	if err != nil {
		return nil, fmt.Errorf("could not get cache directory: %w", err)
	}
	ctx.CacheDir = cacheDir
	ctx.SetCache(caching.NewManager(cacheDir))

	// Initialize cookies
	cookiesDir, err := utils.GetCacheDir("plakar")
	if err != nil {
		return nil, fmt.Errorf("could not get cookies directory: %w", err)
	}
	ctx.SetCookies(cookies.NewManager(cookiesDir))

	// Initialize plugins
	dataDir, err := utils.GetDataDir("plakar")
	if err != nil {
		return nil, fmt.Errorf("could not get data directory: %w", err)
	}
	ctx.SetPlugins(plugins.NewManager(dataDir, cookiesDir))

	// Setup logging
	logger := logging.NewLogger(os.Stdout, os.Stderr)
	if !opts.Quiet {
		logger.EnableInfo()
	}
	if opts.Trace != "" {
		logger.EnableTracing(opts.Trace)
	}
	ctx.SetLogger(logger)

	// Load keyfile if specified
	if opts.KeyFile != "" {
		data, err := os.ReadFile(opts.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("could not read key file: %w", err)
		}
		ctx.KeyFromFile = strings.TrimSuffix(string(data), "\n")
	}

	// Set CPU limit
	runtime.GOMAXPROCS(opts.CPUCount)

	return &ConfigContext{
		AppContext: ctx,
		Options:    opts,
	}, nil
}

// ParseRepository determines the repository path from command line args
func (c *ConfigContext) ParseRepository() error {
	if flag.NArg() == 0 {
		return fmt.Errorf("a subcommand must be provided")
	}

	var repositoryPath string
	var args []string
	var at bool

	if flag.Arg(0) == "at" {
		if len(flag.Args()) < 2 {
			return fmt.Errorf("missing plakar repository")
		}
		if len(flag.Args()) < 3 {
			return fmt.Errorf("missing command")
		}
		repositoryPath = flag.Arg(1)
		args = flag.Args()[2:]
		at = true
	} else {
		repositoryPath = os.Getenv("PLAKAR_REPOSITORY")
		if repositoryPath == "" {
			currentUser, _ := user.Current()
			def := c.Config.DefaultRepository
			if def != "" {
				repositoryPath = "@" + def
			} else {
				repositoryPath = "fs:" + filepath.Join(currentUser.HomeDir, ".plakar")
			}
		}
		args = flag.Args()
	}

	storeConfig, err := c.Config.GetRepository(repositoryPath)
	if err != nil {
		return fmt.Errorf("failed to get repository config: %w", err)
	}

	c.RepositoryPath = repositoryPath
	c.StoreConfig = storeConfig
	c.Args = args
	c.IsAt = at

	return nil
}

// GetPassphraseFromEnv retrieves passphrase from environment or config
func GetPassphraseFromEnv(ctx *appcontext.AppContext, params map[string]string) (string, error) {
	if ctx.KeyFromFile != "" {
		return ctx.KeyFromFile, nil
	}

	if pass, ok := params["passphrase"]; ok {
		delete(params, "passphrase")
		return pass, nil
	}

	if cmd, ok := params["passphrase_cmd"]; ok {
		delete(params, "passphrase_cmd")
		return utils.GetPassphraseFromCommand(cmd)
	}

	if pass, ok := os.LookupEnv("PLAKAR_PASSPHRASE"); ok {
		return pass, nil
	}

	return "", nil
}
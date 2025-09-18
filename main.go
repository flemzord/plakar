package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/PlakarKorp/plakar/internal/bootstrap"

	// Import all subcommands
	_ "github.com/PlakarKorp/plakar/subcommands/agent"
	_ "github.com/PlakarKorp/plakar/subcommands/archive"
	_ "github.com/PlakarKorp/plakar/subcommands/backup"
	_ "github.com/PlakarKorp/plakar/subcommands/cat"
	_ "github.com/PlakarKorp/plakar/subcommands/check"
	_ "github.com/PlakarKorp/plakar/subcommands/clone"
	_ "github.com/PlakarKorp/plakar/subcommands/config"
	_ "github.com/PlakarKorp/plakar/subcommands/create"
	_ "github.com/PlakarKorp/plakar/subcommands/diag"
	_ "github.com/PlakarKorp/plakar/subcommands/diff"
	_ "github.com/PlakarKorp/plakar/subcommands/digest"
	_ "github.com/PlakarKorp/plakar/subcommands/dup"
	_ "github.com/PlakarKorp/plakar/subcommands/help"
	_ "github.com/PlakarKorp/plakar/subcommands/info"
	_ "github.com/PlakarKorp/plakar/subcommands/locate"
	_ "github.com/PlakarKorp/plakar/subcommands/login"
	_ "github.com/PlakarKorp/plakar/subcommands/ls"
	_ "github.com/PlakarKorp/plakar/subcommands/maintenance"
	_ "github.com/PlakarKorp/plakar/subcommands/mount"
	_ "github.com/PlakarKorp/plakar/subcommands/pkg"
	_ "github.com/PlakarKorp/plakar/subcommands/prune"
	_ "github.com/PlakarKorp/plakar/subcommands/ptar"
	_ "github.com/PlakarKorp/plakar/subcommands/restore"
	_ "github.com/PlakarKorp/plakar/subcommands/rm"
	_ "github.com/PlakarKorp/plakar/subcommands/scheduler"
	_ "github.com/PlakarKorp/plakar/subcommands/server"
	_ "github.com/PlakarKorp/plakar/subcommands/service"
	_ "github.com/PlakarKorp/plakar/subcommands/ui"
	_ "github.com/PlakarKorp/plakar/subcommands/version"

	// Import integrations
	_ "github.com/PlakarKorp/integration-fs/exporter"
	_ "github.com/PlakarKorp/integration-fs/importer"
	_ "github.com/PlakarKorp/integration-fs/storage"
	_ "github.com/PlakarKorp/integration-ptar/storage"
	_ "github.com/PlakarKorp/integration-stdio/exporter"
	_ "github.com/PlakarKorp/integration-stdio/importer"
	_ "github.com/PlakarKorp/integration-tar/importer"
)

func entryPoint() int {
	// Override flag.Usage to include command listing
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTIONS] [at REPOSITORY] COMMAND [COMMAND_OPTIONS]...\n", flag.CommandLine.Name())
		fmt.Fprintf(flag.CommandLine.Output(), "\nBy default, the repository is $PLAKAR_REPOSITORY or $HOME/.plakar.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\nOPTIONS:\n")
		flag.PrintDefaults()

		fmt.Fprintf(flag.CommandLine.Output(), "\nCOMMANDS:\n")
		bootstrap.ListCommands(flag.CommandLine.Output(), "  ")
		fmt.Fprintf(flag.CommandLine.Output(), "\nFor more information on a command, use '%s help COMMAND'.\n", flag.CommandLine.Name())
	}

	// Create and configure the bootstrap pipeline
	pipeline := bootstrap.NewPipeline(os.Stdout, os.Stderr)

	// Initialize configuration
	pipeline, err := pipeline.WithConfig()
	if err != nil {
		if flag.NArg() == 0 {
			// Show usage if no command provided
			fmt.Fprintf(os.Stderr, "Error: a subcommand must be provided\n\n")
			bootstrap.ListCommands(os.Stderr, "  ")
		} else {
			log.Printf("Configuration error: %v", err)
		}
		return 1
	}

	// Set up profiling
	pipeline, err = pipeline.WithProfiling()
	if err != nil {
		log.Printf("Profiling setup error: %v", err)
		return 1
	}

	// Handle security checks
	pipeline, err = pipeline.WithSecurity()
	if err != nil {
		// Security flags handled, normal exit
		if err.Error() == "security flag handled, exiting" {
			return 0
		}
		log.Printf("Security check error: %v", err)
		return 1
	}

	// Initialize repository
	pipeline, err = pipeline.WithRepository()
	if err != nil {
		log.Printf("Repository initialization error: %v", err)
		return 1
	}

	// Set up signal handling
	pipeline, err = pipeline.WithSignals()
	if err != nil {
		log.Printf("Signal handler error: %v", err)
		return 1
	}

	// Execute the command
	return pipeline.Execute()
}

func main() {
	os.Exit(entryPoint())
}

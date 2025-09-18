package bootstrap

import (
	"errors"
	"fmt"

	"github.com/PlakarKorp/kloset/encryption"
	"github.com/PlakarKorp/kloset/repository"
	"github.com/PlakarKorp/kloset/storage"
	"github.com/PlakarKorp/kloset/versioning"
	"github.com/PlakarKorp/plakar/subcommands"
	"github.com/PlakarKorp/plakar/utils"
)

var (
	// ErrCantUnlock is returned when the repository cannot be unlocked
	ErrCantUnlock = errors.New("failed to unlock repository")
)

// RepositoryManager handles repository initialization and management
type RepositoryManager struct {
	store      storage.Store
	repository *repository.Repository
	config     *ConfigContext
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(config *ConfigContext) *RepositoryManager {
	return &RepositoryManager{
		config: config,
	}
}

// InitializeRepository initializes the repository based on command flags
func (r *RepositoryManager) InitializeRepository(cmd subcommands.Subcommand, agentless bool) error {
	flags := cmd.GetFlags()

	// Handle commands that don't need repository
	if flags&subcommands.BeforeRepositoryOpen != 0 {
		if r.config.IsAt {
			return fmt.Errorf("command cannot be used with 'at' parameter")
		}
		// store and repo can stay nil
		return nil
	}

	// Handle commands that need storage but not opened repository
	if flags&subcommands.BeforeRepositoryWithStorage != 0 {
		repo, err := repository.Inexistent(r.config.GetInner(), r.config.StoreConfig)
		if err != nil {
			return fmt.Errorf("failed to create inexistent repository: %w", err)
		}
		r.repository = repo
		return nil
	}

	// Open storage and repository
	return r.openRepository(agentless)
}

// openRepository opens the storage and initializes the repository
func (r *RepositoryManager) openRepository(agentless bool) error {
	// Open storage
	store, serializedConfig, err := storage.Open(r.config.GetInner(), r.config.StoreConfig)
	if err != nil {
		location := r.config.StoreConfig["location"]
		return fmt.Errorf("failed to open the repository at %s: %w\nTo specify an alternative repository, please use \"plakar at <location> <command>\"", location, err)
	}
	r.store = store

	// Parse repository configuration
	repoConfig, err := storage.NewConfigurationFromWrappedBytes(serializedConfig)
	if err != nil {
		return fmt.Errorf("failed to parse repository config: %w", err)
	}

	// Check version compatibility
	if repoConfig.Version != versioning.FromString(storage.VERSION) {
		return fmt.Errorf("incompatible repository version: %s != %s", repoConfig.Version, storage.VERSION)
	}

	// Setup encryption if needed
	if err := r.setupEncryption(repoConfig); err != nil {
		return fmt.Errorf("failed to setup encryption: %w", err)
	}

	// Create repository instance
	if agentless {
		repo, err := repository.New(r.config.GetInner(), r.config.GetSecret(), store, serializedConfig)
		if err != nil {
			return fmt.Errorf("failed to create repository: %w", err)
		}
		r.repository = repo
	} else {
		repo, err := repository.NewNoRebuild(r.config.GetInner(), r.config.GetSecret(), store, serializedConfig)
		if err != nil {
			return fmt.Errorf("failed to create repository without rebuild: %w", err)
		}
		r.repository = repo
	}

	return nil
}

// setupEncryption configures encryption for the repository
func (r *RepositoryManager) setupEncryption(config *storage.Configuration) error {
	if config.Encryption == nil {
		return nil
	}

	// Try key from file first
	if r.config.KeyFromFile != "" {
		secret := []byte(r.config.KeyFromFile)
		key, err := encryption.DeriveKey(config.Encryption.KDFParams, secret)
		if err != nil {
			return fmt.Errorf("failed to derive key: %w", err)
		}

		if !encryption.VerifyCanary(config.Encryption, key) {
			return ErrCantUnlock
		}
		r.config.SetSecret(key)
		return nil
	}

	// Fall back to prompting
	for i := 0; i < 3; i++ {
		secret, err := utils.GetPassphrase("repository")
		if err != nil {
			return fmt.Errorf("failed to get passphrase: %w", err)
		}

		key, err := encryption.DeriveKey(config.Encryption.KDFParams, secret)
		if err != nil {
			return fmt.Errorf("failed to derive key: %w", err)
		}

		if encryption.VerifyCanary(config.Encryption, key) {
			r.config.SetSecret(key)
			return nil
		}
	}

	return ErrCantUnlock
}

// GetRepository returns the initialized repository
func (r *RepositoryManager) GetRepository() *repository.Repository {
	return r.repository
}

// GetStore returns the initialized store
func (r *RepositoryManager) GetStore() storage.Store {
	return r.store
}

// Close closes the repository and store
func (r *RepositoryManager) Close() error {
	var errs []error

	if r.repository != nil {
		if err := r.repository.Close(); err != nil {
			errs = append(errs, fmt.Errorf("could not close repository: %w", err))
		}
	}

	if r.store != nil {
		if err := r.store.Close(r.config.AppContext); err != nil {
			errs = append(errs, fmt.Errorf("could not close store: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}
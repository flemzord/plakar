package context

import (
	"context"
	"io"

	"github.com/PlakarKorp/kloset/caching"
	"github.com/PlakarKorp/kloset/logging"
	"github.com/PlakarKorp/plakar/cookies"
	"github.com/PlakarKorp/plakar/plugins"
)

// Provider interface provides access to all context types
type Provider interface {
	// Context access
	Execution() *ExecutionContext
	Security() *SecurityContext
	System() *SystemContext

	// Infrastructure access
	GetLogger() *logging.Logger
	GetCache() *caching.Manager
	GetCookies() *cookies.Manager
	GetPlugins() *plugins.Manager

	// Base context
	Context() context.Context
	Cancel()

	// Cleanup
	Close() error
}

// ContextProvider implements the Provider interface
type ContextProvider struct {
	execution *ExecutionContext
	security  *SecurityContext
	system    *SystemContext

	// Infrastructure components
	logger  *logging.Logger
	cache   *caching.Manager
	cookies *cookies.Manager
	plugins *plugins.Manager

	// Base context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Configuration
	configDir string
	cacheDir  string
	client    string

	// IO
	stdout io.Writer
	stderr io.Writer
}

// NewContextProvider creates a new context provider
func NewContextProvider() *ContextProvider {
	ctx, cancel := context.WithCancel(context.Background())

	return &ContextProvider{
		execution: &ExecutionContext{},
		security:  NewSecurityContext(),
		system:    &SystemContext{},
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Execution returns the execution context
func (p *ContextProvider) Execution() *ExecutionContext {
	return p.execution
}

// Security returns the security context
func (p *ContextProvider) Security() *SecurityContext {
	return p.security
}

// System returns the system context
func (p *ContextProvider) System() *SystemContext {
	return p.system
}

// SetExecutionContext sets the execution context
func (p *ContextProvider) SetExecutionContext(ctx *ExecutionContext) {
	p.execution = ctx
}

// SetSecurityContext sets the security context
func (p *ContextProvider) SetSecurityContext(ctx *SecurityContext) {
	p.security = ctx
}

// SetSystemContext sets the system context
func (p *ContextProvider) SetSystemContext(ctx *SystemContext) {
	p.system = ctx
}

// GetLogger returns the logger
func (p *ContextProvider) GetLogger() *logging.Logger {
	return p.logger
}

// SetLogger sets the logger
func (p *ContextProvider) SetLogger(logger *logging.Logger) {
	p.logger = logger
}

// GetCache returns the cache manager
func (p *ContextProvider) GetCache() *caching.Manager {
	return p.cache
}

// SetCache sets the cache manager
func (p *ContextProvider) SetCache(cache *caching.Manager) {
	p.cache = cache
}

// GetCookies returns the cookies manager
func (p *ContextProvider) GetCookies() *cookies.Manager {
	return p.cookies
}

// SetCookies sets the cookies manager
func (p *ContextProvider) SetCookies(cookies *cookies.Manager) {
	p.cookies = cookies
}

// GetPlugins returns the plugins manager
func (p *ContextProvider) GetPlugins() *plugins.Manager {
	return p.plugins
}

// SetPlugins sets the plugins manager
func (p *ContextProvider) SetPlugins(plugins *plugins.Manager) {
	p.plugins = plugins
}

// Context returns the base context
func (p *ContextProvider) Context() context.Context {
	return p.ctx
}

// Cancel cancels the context
func (p *ContextProvider) Cancel() {
	p.cancel()
}

// SetConfigDir sets the configuration directory
func (p *ContextProvider) SetConfigDir(dir string) {
	p.configDir = dir
}

// GetConfigDir returns the configuration directory
func (p *ContextProvider) GetConfigDir() string {
	return p.configDir
}

// SetCacheDir sets the cache directory
func (p *ContextProvider) SetCacheDir(dir string) {
	p.cacheDir = dir
}

// GetCacheDir returns the cache directory
func (p *ContextProvider) GetCacheDir() string {
	return p.cacheDir
}

// SetClient sets the client string
func (p *ContextProvider) SetClient(client string) {
	p.client = client
}

// GetClient returns the client string
func (p *ContextProvider) GetClient() string {
	return p.client
}

// SetStdout sets the stdout writer
func (p *ContextProvider) SetStdout(w io.Writer) {
	p.stdout = w
}

// GetStdout returns the stdout writer
func (p *ContextProvider) GetStdout() io.Writer {
	return p.stdout
}

// SetStderr sets the stderr writer
func (p *ContextProvider) SetStderr(w io.Writer) {
	p.stderr = w
}

// GetStderr returns the stderr writer
func (p *ContextProvider) GetStderr() io.Writer {
	return p.stderr
}

// Close cleans up all resources
func (p *ContextProvider) Close() error {
	// Clear sensitive data
	if p.security != nil {
		p.security.Clear()
	}

	// Close infrastructure components
	if p.cache != nil {
		p.cache.Close()
	}

	if p.cookies != nil {
		p.cookies.Close()
	}

	// Cancel context
	p.cancel()

	return nil
}
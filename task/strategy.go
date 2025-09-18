package task

import (
	"context"
	"fmt"

	"github.com/PlakarKorp/kloset/objects"
	"github.com/PlakarKorp/kloset/repository"
	"github.com/PlakarKorp/plakar/appcontext"
	"github.com/PlakarKorp/plakar/subcommands"
)

// TaskStrategy interface defines the contract for task execution strategies
type TaskStrategy interface {
	// Kind returns the kind of task (backup, check, restore, etc.)
	Kind() string

	// Execute runs the task and returns status and error
	Execute(ctx context.Context, repo *repository.Repository) (int, error)

	// GetCommand returns the underlying command
	GetCommand() subcommands.Subcommand
}

// TaskResult holds the result of a task execution
type TaskResult struct {
	Status     int
	Error      error
	SnapshotID objects.MAC
	Warning    error
}

// TaskExecutor manages task execution using strategies
type TaskExecutor struct {
	strategies map[string]TaskStrategy
	ctx        *appcontext.AppContext
}

// NewTaskExecutor creates a new task executor
func NewTaskExecutor(ctx *appcontext.AppContext) *TaskExecutor {
	return &TaskExecutor{
		strategies: make(map[string]TaskStrategy),
		ctx:        ctx,
	}
}

// RegisterStrategy registers a task execution strategy
func (e *TaskExecutor) RegisterStrategy(strategy TaskStrategy) {
	e.strategies[strategy.Kind()] = strategy
}

// Execute runs a task using the appropriate strategy
func (e *TaskExecutor) Execute(cmd subcommands.Subcommand, repo *repository.Repository) (*TaskResult, error) {
	// Find the appropriate strategy
	var strategy TaskStrategy
	for _, s := range e.strategies {
		if s.GetCommand() == cmd {
			strategy = s
			break
		}
	}

	if strategy == nil {
		// Fallback to direct execution for commands without strategies
		status, err := cmd.Execute(e.ctx, repo)
		return &TaskResult{
			Status: status,
			Error:  err,
		}, nil
	}

	// Execute using strategy
	status, err := strategy.Execute(e.ctx, repo)
	return &TaskResult{
		Status: status,
		Error:  err,
	}, nil
}

// GetStrategyForCommand returns the strategy for a given command
func (e *TaskExecutor) GetStrategyForCommand(cmd subcommands.Subcommand) TaskStrategy {
	for _, strategy := range e.strategies {
		if strategy.GetCommand() == cmd {
			return strategy
		}
	}
	return nil
}

// BaseTaskStrategy provides common functionality for task strategies
type BaseTaskStrategy struct {
	kind    string
	command subcommands.Subcommand
}

// NewBaseTaskStrategy creates a new base task strategy
func NewBaseTaskStrategy(kind string, cmd subcommands.Subcommand) *BaseTaskStrategy {
	return &BaseTaskStrategy{
		kind:    kind,
		command: cmd,
	}
}

// Kind returns the task kind
func (s *BaseTaskStrategy) Kind() string {
	return s.kind
}

// GetCommand returns the underlying command
func (s *BaseTaskStrategy) GetCommand() subcommands.Subcommand {
	return s.command
}

// Execute provides default execution
func (s *BaseTaskStrategy) Execute(ctx context.Context, repo *repository.Repository) (int, error) {
	if appCtx, ok := ctx.(*appcontext.AppContext); ok {
		return s.command.Execute(appCtx, repo)
	}
	return 1, fmt.Errorf("invalid context type")
}
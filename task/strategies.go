package task

import (
	"context"

	"github.com/PlakarKorp/kloset/repository"
	"github.com/PlakarKorp/plakar/appcontext"
	"github.com/PlakarKorp/plakar/subcommands/backup"
	"github.com/PlakarKorp/plakar/subcommands/check"
	"github.com/PlakarKorp/plakar/subcommands/maintenance"
	"github.com/PlakarKorp/plakar/subcommands/restore"
	"github.com/PlakarKorp/plakar/subcommands/rm"
	"github.com/PlakarKorp/plakar/subcommands/sync"
)

// BackupStrategy handles backup command execution
type BackupStrategy struct {
	*BaseTaskStrategy
	backup *backup.Backup
}

// NewBackupStrategy creates a new backup strategy
func NewBackupStrategy(cmd *backup.Backup) *BackupStrategy {
	return &BackupStrategy{
		BaseTaskStrategy: NewBaseTaskStrategy("backup", cmd),
		backup:          cmd,
	}
}

// Execute runs the backup command
func (s *BackupStrategy) Execute(ctx context.Context, repo *repository.Repository) (int, error) {
	appCtx, ok := ctx.(*appcontext.AppContext)
	if !ok {
		return s.BaseTaskStrategy.Execute(ctx, repo)
	}

	status, err, _, _ := s.backup.DoBackup(appCtx, repo)
	return status, err
}

// CheckStrategy handles check command execution
type CheckStrategy struct {
	*BaseTaskStrategy
	check *check.Check
}

// NewCheckStrategy creates a new check strategy
func NewCheckStrategy(cmd *check.Check) *CheckStrategy {
	return &CheckStrategy{
		BaseTaskStrategy: NewBaseTaskStrategy("check", cmd),
		check:           cmd,
	}
}

// RestoreStrategy handles restore command execution
type RestoreStrategy struct {
	*BaseTaskStrategy
	restore *restore.Restore
}

// NewRestoreStrategy creates a new restore strategy
func NewRestoreStrategy(cmd *restore.Restore) *RestoreStrategy {
	return &RestoreStrategy{
		BaseTaskStrategy: NewBaseTaskStrategy("restore", cmd),
		restore:         cmd,
	}
}

// SyncStrategy handles sync command execution
type SyncStrategy struct {
	*BaseTaskStrategy
	sync *sync.Sync
}

// NewSyncStrategy creates a new sync strategy
func NewSyncStrategy(cmd *sync.Sync) *SyncStrategy {
	return &SyncStrategy{
		BaseTaskStrategy: NewBaseTaskStrategy("sync", cmd),
		sync:            cmd,
	}
}

// RmStrategy handles rm command execution
type RmStrategy struct {
	*BaseTaskStrategy
	rm *rm.Rm
}

// NewRmStrategy creates a new rm strategy
func NewRmStrategy(cmd *rm.Rm) *RmStrategy {
	return &RmStrategy{
		BaseTaskStrategy: NewBaseTaskStrategy("rm", cmd),
		rm:              cmd,
	}
}

// MaintenanceStrategy handles maintenance command execution
type MaintenanceStrategy struct {
	*BaseTaskStrategy
	maintenance *maintenance.Maintenance
}

// NewMaintenanceStrategy creates a new maintenance strategy
func NewMaintenanceStrategy(cmd *maintenance.Maintenance) *MaintenanceStrategy {
	return &MaintenanceStrategy{
		BaseTaskStrategy: NewBaseTaskStrategy("maintenance", cmd),
		maintenance:     cmd,
	}
}

// CreateStrategyForCommand creates the appropriate strategy for a command
func CreateStrategyForCommand(cmd interface{}) TaskStrategy {
	switch c := cmd.(type) {
	case *backup.Backup:
		return NewBackupStrategy(c)
	case *check.Check:
		return NewCheckStrategy(c)
	case *restore.Restore:
		return NewRestoreStrategy(c)
	case *sync.Sync:
		return NewSyncStrategy(c)
	case *rm.Rm:
		return NewRmStrategy(c)
	case *maintenance.Maintenance:
		return NewMaintenanceStrategy(c)
	default:
		return nil
	}
}
package task

import (
	"github.com/PlakarKorp/kloset/objects"
	"github.com/PlakarKorp/kloset/repository"
	"github.com/PlakarKorp/plakar/appcontext"
	"github.com/PlakarKorp/plakar/reporting"
	"github.com/PlakarKorp/plakar/subcommands"
	"github.com/PlakarKorp/plakar/subcommands/backup"
)

func RunCommand(ctx *appcontext.AppContext, cmd subcommands.Subcommand, repo *repository.Repository, taskName string) (int, error) {
	location := ""
	var err error

	if repo != nil {
		location, err = repo.Location()
		if err != nil {
			return 1, err
		}
	}

	reporter := reporting.NewReporter(ctx)
	report := reporter.NewReport()

	// Use strategy pattern instead of type assertions
	strategy := CreateStrategyForCommand(cmd)
	taskKind := ""
	if strategy != nil {
		taskKind = strategy.Kind()
	} else {
		report.SetIgnore()
	}

	report.TaskStart(taskKind, taskName)
	if repo != nil {
		report.WithRepositoryName(location)
		report.WithRepository(repo)
	}

	var status int
	var snapshotID objects.MAC
	var warning error

	// Special handling for backup to get snapshot ID
	if backupCmd, ok := cmd.(*backup.Backup); ok {
		status, err, snapshotID, warning = backupCmd.DoBackup(ctx, repo)
		if !backupCmd.DryRun && err == nil {
			report.WithSnapshotID(snapshotID)
		}
	} else {
		// Use strategy or fallback to direct execution
		if strategy != nil {
			status, err = strategy.Execute(ctx, repo)
		} else {
			status, err = cmd.Execute(ctx, repo)
		}
	}

	if status == 0 {
		if warning != nil {
			report.TaskWarning("warning: %s", warning)
		} else {
			report.TaskDone()
		}
	} else if err != nil {
		report.TaskFailed(0, "error: %s", err)
	}

	reporter.StopAndWait()

	return status, err
}

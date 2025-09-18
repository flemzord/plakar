package task

import (
	"context"
	"testing"

	"github.com/PlakarKorp/kloset/repository"
	"github.com/PlakarKorp/plakar/subcommands"
	"github.com/PlakarKorp/plakar/subcommands/backup"
	"github.com/PlakarKorp/plakar/subcommands/check"
	"github.com/PlakarKorp/plakar/subcommands/maintenance"
	"github.com/PlakarKorp/plakar/subcommands/restore"
	"github.com/PlakarKorp/plakar/subcommands/rm"
	"github.com/PlakarKorp/plakar/subcommands/sync"
)

// Mock command for testing
type mockCommand struct {
	subcommands.SubcommandBase
	executeReturn int
	executeError  error
}

func (m *mockCommand) Parse(ctx interface{}, args []string) error {
	return nil
}

func (m *mockCommand) Execute(ctx interface{}, repo *repository.Repository) (int, error) {
	return m.executeReturn, m.executeError
}

func TestBaseStrategy(t *testing.T) {
	cmd := &mockCommand{
		executeReturn: 0,
		executeError:  nil,
	}

	strategy := &BaseStrategy{
		kind: "test",
		cmd:  cmd,
	}

	// Test Kind()
	if strategy.Kind() != "test" {
		t.Errorf("Expected kind 'test', got '%s'", strategy.Kind())
	}

	// Test GetCommand()
	if strategy.GetCommand() != cmd {
		t.Error("GetCommand() didn't return the correct command")
	}

	// Test Execute()
	status, err := strategy.Execute(context.Background(), nil)
	if status != 0 {
		t.Errorf("Expected status 0, got %d", status)
	}
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCreateStrategyForCommand(t *testing.T) {
	tests := []struct {
		name         string
		command      subcommands.Subcommand
		expectedKind string
		shouldBeNil  bool
	}{
		{
			name:         "Backup command",
			command:      &backup.Backup{},
			expectedKind: "backup",
			shouldBeNil:  false,
		},
		{
			name:         "Check command",
			command:      &check.Check{},
			expectedKind: "check",
			shouldBeNil:  false,
		},
		{
			name:         "Maintenance command",
			command:      &maintenance.Maintenance{},
			expectedKind: "maintenance",
			shouldBeNil:  false,
		},
		{
			name:         "Restore command",
			command:      &restore.Restore{},
			expectedKind: "restore",
			shouldBeNil:  false,
		},
		{
			name:         "Remove command",
			command:      &rm.Remove{},
			expectedKind: "rm",
			shouldBeNil:  false,
		},
		{
			name:         "Sync command",
			command:      &sync.Sync{},
			expectedKind: "sync",
			shouldBeNil:  false,
		},
		{
			name:         "Unknown command",
			command:      &mockCommand{},
			expectedKind: "",
			shouldBeNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := CreateStrategyForCommand(tt.command)

			if tt.shouldBeNil {
				if strategy != nil {
					t.Error("Expected nil strategy for unknown command")
				}
			} else {
				if strategy == nil {
					t.Fatal("Expected non-nil strategy")
				}

				if strategy.Kind() != tt.expectedKind {
					t.Errorf("Expected kind '%s', got '%s'", tt.expectedKind, strategy.Kind())
				}

				if strategy.GetCommand() != tt.command {
					t.Error("Strategy doesn't reference the correct command")
				}
			}
		})
	}
}

func TestBackupStrategy(t *testing.T) {
	cmd := &backup.Backup{}
	strategy := &BackupStrategy{
		BaseStrategy: BaseStrategy{
			kind: "backup",
			cmd:  cmd,
		},
	}

	if strategy.Kind() != "backup" {
		t.Errorf("Expected kind 'backup', got '%s'", strategy.Kind())
	}

	if strategy.GetCommand() != cmd {
		t.Error("GetCommand() didn't return the correct command")
	}
}

func TestCheckStrategy(t *testing.T) {
	cmd := &check.Check{}
	strategy := &CheckStrategy{
		BaseStrategy: BaseStrategy{
			kind: "check",
			cmd:  cmd,
		},
	}

	if strategy.Kind() != "check" {
		t.Errorf("Expected kind 'check', got '%s'", strategy.Kind())
	}

	if strategy.GetCommand() != cmd {
		t.Error("GetCommand() didn't return the correct command")
	}
}

func TestMaintenanceStrategy(t *testing.T) {
	cmd := &maintenance.Maintenance{}
	strategy := &MaintenanceStrategy{
		BaseStrategy: BaseStrategy{
			kind: "maintenance",
			cmd:  cmd,
		},
	}

	if strategy.Kind() != "maintenance" {
		t.Errorf("Expected kind 'maintenance', got '%s'", strategy.Kind())
	}

	if strategy.GetCommand() != cmd {
		t.Error("GetCommand() didn't return the correct command")
	}
}

func TestRestoreStrategy(t *testing.T) {
	cmd := &restore.Restore{}
	strategy := &RestoreStrategy{
		BaseStrategy: BaseStrategy{
			kind: "restore",
			cmd:  cmd,
		},
	}

	if strategy.Kind() != "restore" {
		t.Errorf("Expected kind 'restore', got '%s'", strategy.Kind())
	}

	if strategy.GetCommand() != cmd {
		t.Error("GetCommand() didn't return the correct command")
	}
}

func TestRemoveStrategy(t *testing.T) {
	cmd := &rm.Remove{}
	strategy := &RemoveStrategy{
		BaseStrategy: BaseStrategy{
			kind: "rm",
			cmd:  cmd,
		},
	}

	if strategy.Kind() != "rm" {
		t.Errorf("Expected kind 'rm', got '%s'", strategy.Kind())
	}

	if strategy.GetCommand() != cmd {
		t.Error("GetCommand() didn't return the correct command")
	}
}

func TestSyncStrategy(t *testing.T) {
	cmd := &sync.Sync{}
	strategy := &SyncStrategy{
		BaseStrategy: BaseStrategy{
			kind: "sync",
			cmd:  cmd,
		},
	}

	if strategy.Kind() != "sync" {
		t.Errorf("Expected kind 'sync', got '%s'", strategy.Kind())
	}

	if strategy.GetCommand() != cmd {
		t.Error("GetCommand() didn't return the correct command")
	}
}

func TestStrategyFactory(t *testing.T) {
	// Test that the factory creates the right strategy for each command type
	commands := []struct {
		cmd          subcommands.Subcommand
		expectedKind string
	}{
		{&backup.Backup{}, "backup"},
		{&check.Check{}, "check"},
		{&maintenance.Maintenance{}, "maintenance"},
		{&restore.Restore{}, "restore"},
		{&rm.Remove{}, "rm"},
		{&sync.Sync{}, "sync"},
	}

	for _, tc := range commands {
		strategy := CreateStrategyForCommand(tc.cmd)
		if strategy == nil {
			t.Errorf("Factory returned nil for %T", tc.cmd)
			continue
		}

		if strategy.Kind() != tc.expectedKind {
			t.Errorf("For %T, expected kind '%s', got '%s'",
				tc.cmd, tc.expectedKind, strategy.Kind())
		}
	}
}
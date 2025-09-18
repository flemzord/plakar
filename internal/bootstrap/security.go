package bootstrap

import (
	"fmt"
	"io"
	"os"

	"github.com/PlakarKorp/plakar/appcontext"
	"github.com/PlakarKorp/plakar/utils"
)

// SecurityManager handles security checks and updates
type SecurityManager struct {
	cacheDir              string
	disableSecurityCheck  bool
	stdout                io.Writer
	stderr                io.Writer
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(cacheDir string, stdout, stderr io.Writer) *SecurityManager {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	return &SecurityManager{
		cacheDir: cacheDir,
		stdout:   stdout,
		stderr:   stderr,
	}
}

// HandleSecurityFlags processes security-related command line flags
func (s *SecurityManager) HandleSecurityFlags(ctx *appcontext.AppContext, enableCheck, disableCheck bool) bool {
	if disableCheck {
		ctx.GetCookies().SetDisabledSecurityCheck()
		fmt.Fprintln(s.stdout, "security check disabled !")
		return true
	}

	s.disableSecurityCheck = ctx.GetCookies().IsDisabledSecurityCheck()

	if enableCheck {
		ctx.GetCookies().RemoveDisabledSecurityCheck()
		fmt.Fprintln(s.stdout, "security check enabled !")
		return true
	}

	return false
}

// CheckForUpdates checks for security and reliability updates
func (s *SecurityManager) CheckForUpdates(ctx *appcontext.AppContext) {
	if ctx.GetCookies().IsFirstRun() {
		s.handleFirstRun(ctx)
		return
	}

	if s.disableSecurityCheck {
		return
	}

	// best effort check if security or reliability fix have been issued
	rus, err := utils.CheckUpdate(s.cacheDir)
	if err != nil {
		return
	}

	if !rus.SecurityFix && !rus.ReliabilityFix {
		return
	}

	s.displayUpdateWarning(&rus)
}

// handleFirstRun displays welcome message on first run
func (s *SecurityManager) handleFirstRun(ctx *appcontext.AppContext) {
	ctx.GetCookies().SetFirstRun()

	if s.disableSecurityCheck {
		return
	}

	fmt.Fprintln(s.stdout, "Welcome to plakar !")
	fmt.Fprintln(s.stdout, "")
	fmt.Fprintln(s.stdout, "By default, plakar checks for security updates on the releases feed once every 24h.")
	fmt.Fprintln(s.stdout, "It will notify you if there are important updates that you need to install.")
	fmt.Fprintln(s.stdout, "")
	fmt.Fprintln(s.stdout, "If you prefer to watch yourself, you can disable this permanently by running:")
	fmt.Fprintln(s.stdout, "")
	fmt.Fprintln(s.stdout, "\tplakar -disable-security-check")
	fmt.Fprintln(s.stdout, "")
	fmt.Fprintln(s.stdout, "If you change your mind, run:")
	fmt.Fprintln(s.stdout, "")
	fmt.Fprintln(s.stdout, "\tplakar -enable-security-check")
	fmt.Fprintln(s.stdout, "")
	fmt.Fprintln(s.stdout, "EOT")
}

// displayUpdateWarning shows warning about available updates
func (s *SecurityManager) displayUpdateWarning(rus *utils.ReleaseUpdateSummary) {
	concerns := ""
	if rus.SecurityFix {
		concerns = "security"
	}
	if rus.ReliabilityFix {
		if concerns != "" {
			concerns += " and "
		}
		concerns += "reliability"
	}

	fmt.Fprintf(s.stderr, "WARNING: %s concerns affect your current version, please upgrade to %s (+%d releases).\n",
		concerns, rus.Latest, rus.FoundCount)
}

// IsSecurityCheckDisabled returns true if security checks are disabled
func (s *SecurityManager) IsSecurityCheckDisabled() bool {
	return s.disableSecurityCheck
}
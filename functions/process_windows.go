//go:build windows

package functions

import (
	"os/exec"
	"strconv"
)

// setPlatformProcessAttrs sets platform-specific process attributes.
// On Windows, no special attributes are needed as we use taskkill for tree termination.
func setPlatformProcessAttrs(cmd *exec.Cmd) {
	// No special attributes needed on Windows
}

// killProcessTree kills the process and all its children.
// On Windows, this uses taskkill with /T flag to kill the process tree.
func killProcessTree(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	pid := cmd.Process.Pid

	// Use taskkill to kill the process tree
	// /T = kill child processes
	// /F = force kill
	// /PID = process ID
	killCmd := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid))
	killCmd.Run() // Ignore errors - process might already be dead

	// Also try to kill the process directly as a fallback
	if cmd.Process != nil {
		cmd.Process.Kill()
	}
}

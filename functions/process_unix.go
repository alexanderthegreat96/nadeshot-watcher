//go:build !windows

package functions

import (
	"log"
	"os/exec"
	"syscall"
	"time"
)

// setPlatformProcessAttrs sets platform-specific process attributes.
// On Unix, this creates a new process group so we can kill all children.
func setPlatformProcessAttrs(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killProcessTree kills the process and all its children.
// On Unix, this sends signals to the process group.
func killProcessTree(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	pid := cmd.Process.Pid
	log.Printf("Killing process tree (PID: %d)...", pid)

	// Try SIGINT first - Python handles this better than SIGTERM
	syscall.Kill(-pid, syscall.SIGINT)
	if cmd.Process != nil {
		cmd.Process.Signal(syscall.SIGINT)
	}

	// Give it a moment to handle SIGINT gracefully
	time.Sleep(100 * time.Millisecond)

	// If still alive, try SIGTERM
	if isProcessAlive(pid) {
		log.Printf("Process still alive after SIGINT, sending SIGTERM...")
		syscall.Kill(-pid, syscall.SIGTERM)
		if cmd.Process != nil {
			cmd.Process.Signal(syscall.SIGTERM)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Check if process is still alive and force kill
	if isProcessAlive(pid) {
		log.Printf("Process still alive, sending SIGKILL...")
		syscall.Kill(-pid, syscall.SIGKILL)
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}

	// Wait for process to actually terminate
	waitForProcessDeath(cmd)
}

// isProcessAlive checks if a process with the given PID is still running
func isProcessAlive(pid int) bool {
	// Signal 0 checks if process exists without sending a signal
	err := syscall.Kill(pid, 0)
	return err == nil
}

// waitForProcessDeath waits for the process to actually terminate
func waitForProcessDeath(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	pid := cmd.Process.Pid

	// Wait up to 3 seconds for process to die
	for i := 0; i < 30; i++ {
		if !isProcessAlive(pid) {
			log.Println("Process terminated successfully")
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	log.Println("Warning: Process did not terminate within timeout, forcing kill...")
	syscall.Kill(-pid, syscall.SIGKILL)
	if cmd.Process != nil {
		cmd.Process.Kill()
	}
}

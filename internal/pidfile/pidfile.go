package pidfile

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// Write writes the current process PID to the given path.
func Write(path string) error {
	return os.WriteFile(path, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

// Remove removes the PID file.
func Remove(path string) {
	os.Remove(path)
}

// Read reads the PID from the given file and checks if the process is alive.
// Returns the PID and true if alive, 0 and false otherwise.
func Read(path string) (int, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, false
	}
	// Check if process is alive
	proc, err := os.FindProcess(pid)
	if err != nil {
		return 0, false
	}
	// Signal 0 checks existence without actually sending a signal
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return 0, false
	}
	return pid, true
}

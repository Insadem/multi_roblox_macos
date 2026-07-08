// Package robloxapp contains helpers for spawning, locating, and stopping
// RobloxPlayer.app instances.
package robloxapp

import (
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/Insadem/multi-roblox-macos/pkg/ps"
)

// terminateGrace is how long we give RobloxPlayer to honour SIGTERM
// before escalating to SIGKILL. Roblox installs a SIGTERM handler that
// silently swallows the signal, so we cannot rely on the Signal() return
// value (it only reports whether the kernel delivered the signal, not
// whether the process acted on it). We have to wait and then re-check.
const terminateGrace = 500 * time.Millisecond

// CloseAll iterates over the process table and terminates every process
// whose executable name starts with "RobloxPlayer".
//
// We send SIGTERM first so that Roblox has a chance to flush state to
// disk. If the process is still alive after terminateGrace, we fall back
// to SIGKILL — this matches the user expectation that pressing "close
// all" actually closes everything.
//
// Errors are logged but never returned: this is a best-effort, fire and
// forget operation from the UI.
func CloseAll() {
	// Enumerate twice: processes that spawn between the first snapshot
	// and our SIGTERMs would otherwise survive. A second pass is cheap
	// and covers the common "user clicks the button while a deeplink is
	// still being launched" race.
	for pass := 0; pass < 2; pass++ {
		processes, err := ps.Processes()
		if err != nil {
			log.Printf("closeall: enumerate processes: %v", err)
			return
		}

		matched := false
		for _, process := range processes {
			// kinfoProc.Comm is capped at 16 bytes by the kernel, so a
			// future rename to a longer name would silently truncate.
			// Match by prefix rather than exact equality to stay robust.
			if !strings.HasPrefix(process.Executable(), "RobloxPlayer") {
				continue
			}
			matched = true
			terminate(process.Pid())
		}

		if !matched {
			return
		}
	}
}

// terminate sends SIGTERM to pid, waits up to terminateGrace for the
// process to exit, and escalates to SIGKILL if it is still alive.
func terminate(pid int) {
	p, err := os.FindProcess(pid)
	if err != nil {
		log.Printf("closeall: find pid %d: %v", pid, err)
		return
	}

	// SIGTERM is the "polite quit" signal. macOS GUI apps like
	// RobloxPlayer install a handler for it that does nothing useful, so
	// we cannot rely on the return value of Signal() to know whether the
	// process actually died.
	if err := p.Signal(syscall.SIGTERM); err != nil {
		log.Printf("closeall: SIGTERM pid %d: %v", pid, err)
	}

	// Poll for up to terminateGrace. processHasExited uses kill(pid, 0)
	// to check liveness without sending a signal.
	deadline := time.Now().Add(terminateGrace)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Still alive after the grace period — escalate. RobloxPlayer traps
	// SIGTERM (and SIGINT) but cannot trap SIGKILL, so this always
	// succeeds for processes we own.
	if processAlive(pid) {
		if err := p.Signal(syscall.SIGKILL); err != nil {
			log.Printf("closeall: SIGKILL pid %d: %v", pid, err)
		}
	}
}

// processAlive reports whether pid is still running. We use
// kill(pid, 0), which returns an error with ESRCH if the process is
// gone and nil otherwise. This works on macOS without needing to
// re-enumerate the full process table.
func processAlive(pid int) bool {
	if err := syscall.Kill(pid, 0); err != nil {
		return false
	}
	return true
}

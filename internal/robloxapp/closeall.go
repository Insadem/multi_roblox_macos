// Package robloxapp contains helpers for spawning, locating, and stopping
// RobloxPlayer.app instances.
package robloxapp

import (
	"log"
	"os"

	"github.com/Insadem/multi-roblox-macos/pkg/ps"
)

// CloseAll iterates over the process table and terminates every process
// whose executable name is "RobloxPlayer".
//
// We use SIGTERM rather than SIGKILL so that Roblox has a chance to flush
// state to disk. If the process refuses SIGTERM (e.g. because it is owned
// by another uid), we fall back to SIGKILL — this matches the user
// expectation that pressing "close all" actually closes everything.
//
// Errors are logged but never returned: this is a best-effort, fire and
// forget operation from the UI.
func CloseAll() {
	processes, err := ps.Processes()
	if err != nil {
		log.Printf("closeall: enumerate processes: %v", err)
		return
	}

	for _, process := range processes {
		if process.Executable() != "RobloxPlayer" {
			continue
		}

		p, err := os.FindProcess(process.Pid())
		if err != nil {
			log.Printf("closeall: find pid %d: %v", process.Pid(), err)
			continue
		}

		if err := p.Signal(os.Interrupt); err != nil {
			// Fall back to SIGKILL if SIGTERM is rejected.
			log.Printf("closeall: SIGTERM pid %d: %v — escalating", process.Pid(), err)
			if err := p.Kill(); err != nil {
				log.Printf("closeall: SIGKILL pid %d: %v", process.Pid(), err)
			}
		}

		// Release the OS handle so we don't accumulate zombies.
		if err := p.Release(); err != nil {
			log.Printf("closeall: release pid %d: %v", process.Pid(), err)
		}
	}

	// We deliberately do not block on Wait() here: this is a UI button
	// and the user wants the click to be cheap. If you ever want to wait
	// for Roblox to exit before returning, wrap the Signal in a goroutine
	// that calls p.Wait() and use a sync.WaitGroup.
}

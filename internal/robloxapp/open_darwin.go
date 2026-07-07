//go:build darwin
// +build darwin

package robloxapp

import (
	"os"
	"os/exec"
)

// Open launches /Applications/Roblox.app's main executable directly,
// bypassing Launch Services. This is only useful for tests and small
// command-line tools; production code should go through exec.Command("open", ...).
//
// The returned cleanup function sends SIGTERM and then SIGKILL if the
// process is still running, matching CloseAll's behaviour. We do not call
// Wait() so the caller controls timing.
func Open() (func(), error) {
	cmd := exec.Command("/Applications/Roblox.app/Contents/MacOS/RobloxPlayer")
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return func() {
		if cmd.Process == nil {
			return
		}
		// Try SIGTERM first to give Roblox a chance to clean up.
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			// Fall back to SIGKILL.
			_ = cmd.Process.Kill()
		}
		_ = cmd.Process.Release()
	}, nil
}

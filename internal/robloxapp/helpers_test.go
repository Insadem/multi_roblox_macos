package robloxapp

import (
	"os"
	"testing"
)

// requireRoblox skips the calling test if /Applications/Roblox.app is
// not present. The integration tests in this package copy and launch
// the installed Roblox bundle; without it, the assertions are
// meaningless. CI runs on macos-14 with Roblox preinstalled, so the
// skip rarely fires there; on a developer machine without Roblox the
// tests are skipped cleanly instead of erroring.
func requireRoblox(t *testing.T) {
	t.Helper()
	if _, err := os.Stat("/Applications/Roblox.app"); err != nil {
		t.Skip("/Applications/Roblox.app not installed; skipping integration test")
	}
}

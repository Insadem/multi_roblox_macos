package syncbreaker

import (
	"os"
	"testing"
)

// requireRoblox skips the calling test if /Applications/Roblox.app is
// not present. TestBreak depends on a real Roblox launch creating the
// /RobloxPlayerUniq semaphore; without it the assertion is
// meaningless. CI runs on macos-14 with Roblox preinstalled, so the
// skip rarely fires there.
func requireRoblox(t *testing.T) {
	t.Helper()
	if _, err := os.Stat("/Applications/Roblox.app"); err != nil {
		t.Skip("/Applications/Roblox.app not installed; skipping integration test")
	}
}

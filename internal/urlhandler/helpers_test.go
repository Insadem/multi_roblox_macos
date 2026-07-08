package urlhandler

import (
	"os"
	"testing"
)

// requireRoblox skips the calling test if /Applications/Roblox.app
// is not present. TestUrlHandler depends on the system having a
// Roblox bundle to register as the default handler; without it the
// assertions are meaningless. CI runs on macos-14 with Roblox
// preinstalled, so the skip rarely fires there.
func requireRoblox(t *testing.T) {
	t.Helper()
	if _, err := os.Stat("/Applications/Roblox.app"); err != nil {
		t.Skip("/Applications/Roblox.app not installed; skipping integration test")
	}
}

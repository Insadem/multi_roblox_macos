package urlhandler

import "testing"

func TestUrlHandler(t *testing.T) {
	requireRoblox(t)

	ROBLOX_BUNDLE_IDENTIFIER := "com.roblox.RobloxPlayer"
	v := Check(ROBLOX_BUNDLE_IDENTIFIER, "roblox")
	if v != true {
		t.Error("roblox handler supposed to be set to default")
	}

	v = Check(ROBLOX_BUNDLE_IDENTIFIER, "roblox-player")
	if v != true {
		t.Error("roblox-player handler supposed to be set to default")
	}

	ok := Set("com.Roblox.RobloxStudio", "roblox-player")
	if !ok {
		t.Error("set expected to be ok")
	}

	v = Check(ROBLOX_BUNDLE_IDENTIFIER, "roblox-player")
	if v == true {
		t.Error("roblox-player handler supposed to be not set")
	}

	Set(ROBLOX_BUNDLE_IDENTIFIER, "roblox-player")
	v = Check(ROBLOX_BUNDLE_IDENTIFIER, "roblox-player")
	if v != true {
		t.Error("roblox-player handler supposed to be set to default")
	}

	v = Check(ROBLOX_BUNDLE_IDENTIFIER, "non-existant")
	if v == true {
		t.Error("supposed to be not set")
	}
}

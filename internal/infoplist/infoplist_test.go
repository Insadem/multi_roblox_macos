package infoplist

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleInfoPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleIdentifier</key>
	<string>com.roblox.RobloxPlayer</string>
	<key>LSMultipleInstancesProhibited</key>
	<true/>
</dict>
</plist>
`

// TestSetMultipleInstancesProhibition is a pure unit test: it writes a
// representative Info.plist into a temp directory, runs the modifier,
// then re-reads the result and asserts the key was flipped. This avoids
// the original test's habit of editing /Applications/Roblox.app, which
// is destructive and breaks in CI.
func TestSetMultipleInstancesProhibition(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Info.plist")

	if err := os.WriteFile(path, []byte(sampleInfoPlist), 0o644); err != nil {
		t.Fatalf("seed plist: %v", err)
	}

	if err := SetMultipleInstancesProhibition(path, false); err != nil {
		t.Fatalf("set: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}

	// Quick textual check — a full re-parse via howett.net/plist is
	// also possible but it makes the test depend on the same library
	// we are testing, which is fragile.
	if want := "<false/>"; !contains(got, want) {
		t.Errorf("expected plist to contain %q, got:\n%s", want, got)
	}
	if !contains(got, "<key>CFBundleIdentifier</key>") {
		t.Errorf("expected plist to still contain CFBundleIdentifier, got:\n%s", got)
	}
}

// TestSetMultipleInstancesProhibition_MissingFile verifies we return a
// useful error when the plist does not exist, rather than panicking.
func TestSetMultipleInstancesProhibition_MissingFile(t *testing.T) {
	err := SetMultipleInstancesProhibition("/no/such/file.plist", true)
	if err == nil {
		t.Fatal("expected error for missing plist")
	}
}

func contains(haystack []byte, needle string) bool {
	if len(needle) > len(haystack) {
		return false
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

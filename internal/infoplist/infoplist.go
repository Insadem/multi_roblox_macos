// Package infoplist edits Info.plist files of .app bundles.
//
// The only knob it currently exposes is LSMultipleInstancesProhibited, which
// Roblox uses to refuse launching a second copy of the player. Setting it to
// false on a per-copy Info.plist is the trick that lets a second instance
// run side-by-side with the original.
package infoplist

import (
	"bytes"
	"fmt"
	"os"

	"howett.net/plist"
)

// SetMultipleInstancesProhibition rewrites the LSMultipleInstancesProhibited
// key in the Info.plist at path. It preserves the file's original XML/binary
// encoding so that macOS code-signing and any cached parsers see a minimal
// change.
//
// path is the full path to an Info.plist inside a .app bundle, e.g.
// /Applications/Roblox.app/Contents/Info.plist.
func SetMultipleInstancesProhibition(path string, prohibited bool) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open plist: %w", err)
	}
	defer file.Close()

	dec := plist.NewDecoder(file)

	var val map[string]interface{}
	if err := dec.Decode(&val); err != nil {
		return fmt.Errorf("decode plist: %w", err)
	}
	if val == nil {
		val = make(map[string]interface{})
	}

	val["LSMultipleInstancesProhibited"] = prohibited

	buf := &bytes.Buffer{}
	encoder := plist.NewEncoderForFormat(buf, dec.Format)
	if err := encoder.Encode(val); err != nil {
		return fmt.Errorf("encode plist: %w", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write plist: %w", err)
	}

	return nil
}

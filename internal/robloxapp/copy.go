package robloxapp

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/Insadem/multi-roblox-macos/pkg/fspath"
)

// Copy is a freshly-cloned /Applications/Roblox.app sitting in the system
// temp directory. It implements io.Closer so the caller can delete it via
// the standard pattern (defer copy.Close()).
type Copy struct {
	dir string
}

// Path returns the absolute path to the cloned .app bundle.
func (b Copy) Path() string {
	return b.dir + "/Roblox.app"
}

// Close releases all resources associated with the copy by recursively
// deleting the temp directory. It is safe to call Close more than once.
func (b Copy) Close() error {
	if err := os.RemoveAll(b.dir); err != nil {
		return fmt.Errorf("remove roblox app copy: %w", err)
	}
	return nil
}

// Compile-time assertion that Copy satisfies io.Closer.
var _ io.Closer = (*Copy)(nil)

// copyDestDir makes a uniquely-named subdirectory inside the system temp
// directory and returns its absolute path.
//
// We try MkdirTemp first (atomic, race-free). If that fails for any
// reason, we fall back to Mkdir + a manual uniqueness suffix so the
// caller still gets a usable directory.
//
// Errors from the fallback are returned to the caller. The function is
// the only place where we create a parent dir, so the result is safe to
// use as a destination for `cp -a` and for os.RemoveAll on Close.
func copyDestDir() (string, error) {
	d, err := fspath.TMPDir.Get()
	if err != nil {
		return "", fmt.Errorf("locate temp dir: %w", err)
	}

	// MkdirTemp atomically creates a uniquely-named directory.
	path, err := os.MkdirTemp(d, "multi-roblox-")
	if err != nil {
		// On the rare chance MkdirTemp fails (e.g. weird permission
		// edge cases), fall back to a manual unique directory.
		path, err2 := fallbackUniqueDir(d)
		if err2 != nil {
			return "", fmt.Errorf("mkdtemp %s: %v (fallback also failed: %v)", d, err, err2)
		}
		return path, nil
	}

	return path, nil
}

// fallbackUniqueDir makes a directory by appending a counter to the prefix
// until Mkdir succeeds. Used only when os.MkdirTemp fails.
func fallbackUniqueDir(parent string) (string, error) {
	for i := 0; i < 1<<10; i++ {
		path := fmt.Sprintf("%s/multi-roblox-%d", parent, i)
		if err := os.Mkdir(path, 0o700); err == nil {
			return path, nil
		} else if !errors.Is(err, fs.ErrExist) {
			return "", err
		}
	}
	return "", errors.New("exhausted fallback attempts to make a unique temp dir")
}

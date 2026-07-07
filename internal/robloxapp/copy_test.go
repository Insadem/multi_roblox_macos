package robloxapp

import (
	"os"
	"testing"
)

func TestCopy(t *testing.T) {
	requireRoblox(t)

	copy, err := NewCopy()
	if err != nil {
		t.Error(err)
		return
	}

	path := copy.Path()
	if _, err := os.Stat(path); err != nil {
		t.Error(err)
		return
	}

	err = copy.Close()
	if err != nil {
		t.Error(err)
		return
	}

	if _, err := os.Stat(path); err == nil {
		t.Error("Roblox app still exists")
		return
	}
}

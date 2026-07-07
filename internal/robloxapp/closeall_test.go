package robloxapp

import (
	"testing"
	"time"
)

func TestCloseAll(t *testing.T) {
	requireRoblox(t)

	close, err := Open()
	if err != nil {
		t.Error(err)
	}
	defer close()

	time.Sleep(time.Millisecond * 250)
	CloseAll()
}

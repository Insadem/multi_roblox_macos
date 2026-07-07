package syncbreaker

import (
	"testing"
	"time"

	"github.com/Insadem/multi-roblox-macos/internal/robloxapp"
)

func TestBreak(t *testing.T) {
	requireRoblox(t)

	close, err := robloxapp.Open()
	if err != nil {
		t.Error(err)
	}
	defer close()
	time.Sleep(time.Millisecond * 666)

	ok := Break()
	if !ok {
		t.Error("can't destroy semaphore")
	}

	ok = Break()
	if ok {
		t.Error("expected to be not able destroy semaphore")
	}
}

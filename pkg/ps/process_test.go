package ps

import (
	"os"
	"testing"
)

func TestFindProcess(t *testing.T) {
	p, err := FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if p == nil {
		t.Fatal("should have process")
	}

	if p.Pid() != os.Getpid() {
		t.Fatalf("bad: %#v", p.Pid())
	}
}

func TestProcesses(t *testing.T) {
	// This test works because there will always be SOME processes
	// running. We just assert the list is non-empty and contains at
	// least one process whose name we can match against the test
	// runner's own pid. This is more robust than hard-coding a binary
	// name like "gopls" or "go.exe" that may not exist on every
	// developer's machine.
	p, err := Processes()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(p) <= 0 {
		t.Fatal("should have processes")
	}

	// FindProcess should be able to look us up by pid, which proves
	// the iteration above covers our own process.
	self, err := FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("find self: %s", err)
	}
	if self == nil {
		t.Fatal("expected to find own process in process list")
	}
}

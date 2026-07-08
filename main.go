// Command multi-roblox-macos is a tiny macOS launcher that lets you run
// several RobloxPlayer.app copies in parallel from the same machine.
//
// Roblox installs a single RobloxPlayer.app and its Info.plist declares
// LSMultipleInstancesProhibited = true, which makes Launch Services refuse
// to start a second instance. To get around that, this program:
//
//  1. Registers itself (com.insadem.multiroblox) as the default handler for
//     the roblox-player:// URL scheme, so the browser hands us every "play"
//     click on roblox.com.
//  2. On each deeplink it makes a fresh copy of /Applications/Roblox.app in
//     the system temp dir, flips LSMultipleInstancesProhibited off in the
//     copy's Info.plist, and opens it.
//  3. Keeps a list of all copies it has spawned; on quit it deletes them.
package main

import (
	"io"
	"log"
	"os/exec"

	"github.com/Insadem/multi-roblox-macos/internal/deeplink"
	"github.com/Insadem/multi-roblox-macos/internal/infoplist"
	"github.com/Insadem/multi-roblox-macos/internal/macosapp"
	"github.com/Insadem/multi-roblox-macos/internal/robloxapp"
	"github.com/Insadem/multi-roblox-macos/internal/syncbreaker"
	"github.com/Insadem/multi-roblox-macos/internal/urlhandler"
)

func main() {
	// Take over the roblox-player:// scheme. If that fails, warn the user
	// but keep running so they can still see the existing copies' status.
	if !urlhandler.Set("com.insadem.multiroblox", "roblox-player") {
		macosapp.Notification("could not register as roblox-player:// handler — open the app manually first")
	}
	if urlhandler.Check(urlhandler.ROBLOX_BUNDLE_IDENTIFIER, "roblox-player") {
		// Roblox still owns the scheme. We re-register on every deeplink
		// below, and the user can re-launch this app to retry.
		log.Println("roblox-player:// is currently owned by Roblox; will retry on first deeplink")
	}

	// Buffer the channel so the deeplink goroutine never blocks on send.
	// On shutdown we drain it.
	copies := make(chan io.Closer, 32)
	urls := deeplink.Handler()

	// stop is closed when the UI is asked to quit. The deeplink goroutine
	// selects on it; the cleanup callback also uses it as a signal that
	// it's safe to drain and exit.
	stop := make(chan struct{})

	// deeplink goroutine: spawns a fresh Roblox.app copy for every URL the
	// browser hands us.
	go func() {
		for {
			select {
			case <-stop:
				return
			case url, ok := <-urls:
				if !ok {
					return
				}
				spawnCopy(url, copies, stop)
			}
		}
	}()

	macosapp.Run(func() {
		// On quit, restore Roblox as the roblox-player:// handler so the
		// system is left in a sensible state.
		urlhandler.Set(urlhandler.ROBLOX_BUNDLE_IDENTIFIER, "roblox-player")

		// Tell the deeplink goroutine to stop, then drain anything it
		// already produced. The channel is buffered (see above) so the
		// goroutine can return even if we don't drain immediately.
		close(stop)
		for {
			select {
			case c := <-copies:
				_ = c.Close()
			default:
				return
			}
		}
	})
}

// spawnCopy produces one Roblox.app copy and posts it to copies.
//
// All error paths close the copy before returning so we never leak temp
// directories when something goes wrong.
func spawnCopy(url string, copies chan<- io.Closer, stop <-chan struct{}) {
	copy, err := robloxapp.NewCopy()
	if err != nil {
		macosapp.Notification("can't make new app copy: " + err.Error())
		return
	}

	// Hand the copy to the shutdown handler. If we are already shutting
	// down, dispose of it and bail.
	select {
	case copies <- copy:
	case <-stop:
		_ = copy.Close()
		return
	}

	// Roblox uses a named semaphore ("/RobloxPlayerUniq") to detect an
	// already-running player. Destroying it before launching our copy is
	// what lets the new instance start without seeing the old one.
	if !syncbreaker.Break() {
		// If we can't destroy the semaphore, the copy will likely fail
		// to start. Tell the user.
		macosapp.Notification("warning: could not unlink /RobloxPlayerUniq — the new instance may refuse to start")
	}

	if err := exec.Command("open", "-a", copy.Path(), url).Run(); err != nil {
		macosapp.Notification("can't open roblox app copy: " + err.Error())
		_ = copy.Close()
		return
	}

	// Flip LSMultipleInstancesProhibited on the copy. This guarantees
	// the copy is not blocked by Launch Services before the app binary
	// is launched. (Some copies are pre-patched; this keeps the project
	// resilient against Roblox changing their plist.)
	if err := infoplist.SetMultipleInstancesProhibition(copy.Path()+"/Contents/Info.plist", false); err != nil {
		macosapp.Notification("can't set multi instance prohibition: " + err.Error())
		return
	}

	// Re-break the semaphore in case Roblox re-created it while we were
	// launching the copy.
	syncbreaker.Break()
}

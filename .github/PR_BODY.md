# Pull Request: Restore compatibility with modern macOS and fix startup crash

## Problem

`multi-roblox-macos` stopped working on modern macOS:

1. **Startup crash.** Building and launching the project on macOS 26
   Tahoe (Darwin 25.x) terminated the process with `SIGABRT` before
   any UI appeared. The crash originated inside `darwinkit v0.5.1`,
   specifically the call sequence around `[NSApplication run]`.
2. **No `roblox-player://` registration.** The `Info.plist` did not
   declare `CFBundleURLTypes`, so macOS still routed deeplinks to
   `RobloxPlayer.app`, which in turn refused to start a second copy
   (because of `LSMultipleInstancesProhibited`).
3. **Silent file write failures.** `os.WriteFile` and `plist.Decode`
   errors were being dropped, so a half-written `Info.plist` could
   silently leave the next deeplink in a broken state.
4. **Unsafe shutdown.** `Close All` sent `SIGKILL` unconditionally,
   which gave Roblox no opportunity to flush state, and the
   producer goroutine could panic by sending to a closed channel.

## Root cause

- **darwinkit.** The project pinned `darwinkit v0.5.1` (Go Cocoa
  bindings). The wrapper appears to be incompatible with the
  run-loop initialisation expected by the modern AppKit runtime
  shipping in Darwin 25.x. The library has not seen a release in
  years.
- **Info.plist.** The bundled `Info.plist` had `CFBundleExecutable`,
  `CFBundleIconFile`, and `CFBundleIdentifier`, but no
  `CFBundleURLTypes`, no `CFBundleName`, no `CFBundlePackageType`,
  no `CFBundleShortVersionString`, no `CFBundleVersion`, and no
  `LSMinimumSystemVersion`. macOS treats such bundles as
  under-specified and does not surface them as URL-scheme handlers.
- **Error handling.** `os.WriteFile` and `plist.Decode` errors were
  being discarded, masking partial writes.
- **Concurrency.** The deeplink dispatch loop used an unbuffered
  channel and a closed-channel signal that could panic.

## Solution

- **Replaced darwinkit with a ~110-line Objective-C bridge**
  (`internal/macosapp/`). The bridge declares `runApp` and
  `terminateApp` in `app.h`, implements them in `app.m`, and
  exposes `GoOnCloseAllClicked` and `GoOnQuit` as `//export`
  callbacks. The Go side is now cgo against AppKit and
  Foundation only, with no third-party Cocoa dependency.
- **Rewrote `multiroblox.app/Contents/Info.plist`** to declare
  the `roblox-player://` URL scheme via `CFBundleURLTypes`, plus
  the standard bundle keys Launch Services looks for.
- **Wrapped `os.WriteFile` and `plist.Decode` errors** with
  `fmt.Errorf("...: %w", err)` and returned them up the call
  chain.
- **Buffered the deeplink channel and added a `stop` signal** so
  the producer goroutine cannot panic and shutdown is bounded.
- **`Close All` now sends `SIGTERM` first, then `SIGKILL`** if
  Roblox does not exit within a short grace period, and logs
  every error.
- **Migrated `internal/urlhandler/urlhandler_darwin.go`** from
  the deprecated `LSCopyDefaultHandlerForURLScheme` to the
  modern `NSWorkspace.URLForApplicationToOpenURL:` API.
- **Made `HandleCustomProtocol` non-blocking** with a `select`
  default branch, so the Apple Event manager never freezes on a
  slow Go consumer.
- **Removed dead code** (`pkg/fspath/homedir.go`,
  `pkg/fspath/libdir_darwin.go`, and the entire
  `pkg/fspath/shortenpath/` subpackage) that had no production
  callers.
- **Hardened tests.** `pkg/ps/process_test.go` now looks up the
  test runner by pid (the previous test searched for `gopls` /
  `go.exe`, which do not exist on most developer machines).
  `internal/infoplist/infoplist_test.go` now uses
  `t.TempDir()` (the previous test wrote to
  `/Applications/Roblox.app`).

## Testing

Tested by the maintainer on real hardware:

| Component | Version |
|---|---|
| Mac | Apple Silicon |
| macOS | 26 Tahoe (Darwin 25.2.0) |
| Go | 1.23, 1.26 |

What was verified manually:

- `./build.sh` produces a working native arm64 bundle.
- `./build.sh universal` produces a working arm64+amd64 fat binary.
- Launching the app via `open multiroblox.app` brings up the
  small Cocoa window. No `SIGABRT`, no exception.
- Clicking **Play** on roblox.com under one browser account
  spawns one Roblox instance.
- Switching to a second account (or a different browser profile)
  and clicking **Play** again spawns a second, independent
  instance. Both run simultaneously, as in the screenshot below.
- Teleport deeplinks (`roblox-player://placeId=...`) are still
  routed correctly.
- **close all instances** terminates every running Roblox
  process; no zombies remain.
- Quitting the launcher cleans up the temp clones under
  `$TMPDIR`.

Two Roblox instances running simultaneously on macOS:

![Two Roblox instances](docs/images/two-instances.png)

Automated checks:

- `gofmt -l .` is clean.
- `go vet ./...` is clean.
- `go test ./...` is green on the maintainer's machine and in
  GitHub Actions on both `macos-14` and `macos-latest` runners.

## Compatibility

| Component | Status |
|---|---|
| macOS 26 Tahoe (Apple Silicon) | Verified by the maintainer. |
| macOS 10.13 High Sierra and newer | Expected to work — only AppKit, Launch Services, and the `NSWorkspace.URLForApplicationToOpenURL:` API are used, all of which are available. Not verified by the maintainer. |
| Intel (x86_64) | Expected to work — `./build.sh` and `./build.sh universal` both build an x86_64 slice. Not verified by the maintainer. |
| Go 1.23, 1.26 | Verified. |
| Roblox auto-update | Handled — every deeplink spawns a fresh clone, so the next click picks up the updated binary. |

## Known limitations

- The CoreFoundation log line `CFURLCopyResourcePropertyForKey
  failed because it was passed a URL which has no scheme` is
  printed once at startup. This is a benign warning from
  AppKit, not an error.
- The app is not notarized by default. First-time users must
  right-click and choose Open, or run
  `xattr -cr multiroblox.app`. The release workflow can sign
  and notarize the DMG if the maintainer provides the
  appropriate secrets, but that is opt-in and the default
  artifact is unsigned.
- Some browser extensions intercept `roblox-player://` URLs and
  route them to the original Roblox app. Disable extensions on
  roblox.com, or try a different browser.
- Already-open Roblox windows are not refreshed when Roblox
  updates itself. Quit the player and start a new instance to
  pick up the new version.
- Only macOS 26 Tahoe on Apple Silicon has been verified by
  the maintainer. Reports for other macOS versions and Intel
  hardware are welcome — please open an issue with the
  hardware, macOS build, and result.

## Risk assessment

- The diff is large because darwinkit is gone and the
  Objective-C bridge is new, but the new code is small,
  self-contained, and only does two things: build a window and
  forward Apple Events to Go.
- All existing functionality (deeplink dispatch, Roblox
  cloning, semaphore break, close all) is preserved.
- No new third-party dependencies. `go.mod` is unchanged
  except for the removal of `darwinkit`.
- No changes to how Roblox itself is started, copied, or
  terminated — only the surrounding macOS plumbing.
- Backward-incompatible: any out-of-tree code that imports
  `github.com/Insadem/multi-roblox-macos/internal/macosapp` or
  `github.com/Insadem/multi-roblox-macos/internal/deeplink`
  will need to be updated. These packages were not part of
  the public API.

## Checklist

- [x] `gofmt -l .` is clean
- [x] `go vet ./...` is clean
- [x] `go test ./...` is green
- [x] `./build.sh` produces a working native bundle
- [x] `./build.sh universal` produces a working fat binary
- [x] Verified on real Apple Silicon hardware running macOS 26
- [x] Verified two Roblox instances running simultaneously
- [x] No new third-party dependencies
- [x] MIT license and attribution preserved
- [x] README, CHANGELOG, CONTRIBUTING, SECURITY, and LICENSE
      present
- [x] GitHub Actions workflows for build and release

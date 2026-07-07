# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.2.0] - 2026-07-07

### Added
- Native Cocoa UI written in Objective-C, replacing the darwinkit dependency.
- `CFBundleURLTypes` and standard bundle keys (`CFBundleName`, `CFBundlePackageType`,
  `CFBundleShortVersionString`, `CFBundleVersion`, `LSMinimumSystemVersion`,
  `NSHighResolutionCapable`) in `multiroblox.app/Contents/Info.plist`, so
  macOS now correctly identifies the bundle and routes `roblox-player://`
  deeplinks to it.
- `build.sh` — convenience script that compiles the binary into
  `multiroblox.app/Contents/MacOS` and supports a `universal` flag.
- README, CHANGELOG, CONTRIBUTING, SECURITY, and LICENSE files for
  professional open-source release.
- GitHub Actions workflows for native, Intel, and universal builds plus
  signed DMG release.
- Universal binary support (arm64 + amd64) via `lipo`.

### Fixed
- **Startup crash on macOS 15+ (Darwin 25.x).** `darwinkit v0.5.1` calls
  `[NSApplication run]` in a way that throws a C++ exception on modern
  Cocoa runtimes, aborting the process before any UI appears. Replaced
  with a minimal Objective-C bridge.
- `internal/infoplist/infoplist.go` — `os.WriteFile` and `plist.Decode`
  errors were being silently dropped, leading to half-written Info.plist
  files. Both errors are now wrapped and returned.
- `main.go` — producer goroutine could panic by sending to a closed
  channel during shutdown, and the deeplink URL handler was only
  registered on the first launch. The dispatch loop is now buffered
  and uses a `stop` signal for clean termination.
- `internal/robloxapp/closeall.go` — was sending SIGKILL unconditionally.
  Now sends SIGTERM first, falls back to SIGKILL if rejected, logs every
  error, and releases OS handles to avoid zombie processes.
- `internal/robloxapp/copy.go` — `os.MkdirTemp` had no fallback path.
  Added a counter-based fallback so the launcher still works in
  permission-edge-case scenarios.
- `internal/deeplink/deeplink.go` — `HandleCustomProtocol` was a blocking
  send, which could freeze the Apple Event manager if the Go consumer
  was slow. Now uses a non-blocking send.
- `internal/urlhandler/urlhandler_darwin.go` — used the deprecated
  `LSCopyDefaultHandlerForURLScheme`. Migrated to the modern
  `NSWorkspace.URLForApplicationToOpenURL` API.
- `pkg/ps/process_test.go` — original test searched for `gopls` or
  `go.exe`, which do not exist on most developer machines. Now looks up
  the test runner itself by pid.
- `internal/infoplist/infoplist_test.go` — original test was destructive
  (wrote to `/Applications/Roblox.app`). Now uses a temp directory.

### Changed
- Dead code removal: `pkg/fspath/homedir.go`, `pkg/fspath/libdir_darwin.go`,
  and the entire `pkg/fspath/shortenpath/` subpackage. None had any
  production callers.
- Module name kept as `github.com/Insadem/multi-roblox-macos` for
  compatibility with the upstream repository.

### Known Issues
- The CoreFoundation log line `CFURLCopyResourcePropertyForKey failed
  because it was passed a URL which has no scheme` is printed once at
  startup. This is a benign warning from AppKit, not an error; the
  window appears normally and the app functions correctly.
- Roblox updates invalidate the running copy. The launcher picks up
  the new binary automatically on the next deeplink, but already-open
  Roblox windows will keep running the old version until you close
  them.
- The app is not notarized. First-time users must right-click and
  choose Open, or run `xattr -cr multiroblox.app` to strip the
  quarantine attribute. A signed and notarized DMG is available in
  the GitHub release.

## [2.0.0] - 2024

Original release by [Insadem](https://github.com/Insadem). Multi-instance
support via deeplink dispatch and `/Applications/Roblox.app` cloning,
implemented with the darwinkit Cocoa wrapper.

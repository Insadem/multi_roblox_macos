# Contributing to multi-roblox-macos

Thank you for your interest in improving this project. Bug reports, feature
proposals, documentation fixes, and pull requests are all welcome.

## Code of conduct

Be respectful. Disagree on substance, not on people. Assume good faith.
Harassment of any kind is not tolerated.

## How to contribute

### Reporting bugs

Open a [GitHub issue](https://github.com/kushangshah/multi-roblox-macos/issues)
and include:

- The exact macOS version (`sw_vers` output).
- The exact Go version if you built from source (`go version`).
- The exact Roblox version, if relevant.
- A minimal reproduction: what you clicked, what you expected, what happened.
- Relevant log output. Run the binary directly from a terminal to capture it:
  ```sh
  ./multiroblox.app/Contents/MacOS/multi-roblox-macos
  ```

### Proposing features

Open an issue first. Discuss the design before you write code; the project
is small enough that an afternoon of discussion saves a week of refactoring.

### Submitting a pull request

1. Fork the repository and create a topic branch off `main`:
   ```sh
   git checkout -b fix/short-description
   ```
2. Make your change. Follow the existing style:
   - Run `gofmt` on every modified Go file.
   - Run `go vet ./...`.
   - Add a test in `*_test.go` next to the code you changed.
3. Update the `CHANGELOG.md` under `[Unreleased]` if the change is
   user-visible.
4. Commit using [Conventional Commits](https://www.conventionalcommits.org/):
   - `fix: <description>` for bug fixes
   - `feat: <description>` for new features
   - `docs: <description>` for documentation only
   - `refactor: <description>` for refactors with no behavior change
   - `test: <description>` for test-only changes
   - `chore: <description>` for build, CI, or tooling
5. Push your branch and open a pull request against `main`.
6. Make sure the GitHub Actions CI is green.

## Development setup

### Prerequisites

- macOS 11 or newer
- Go 1.23 or newer
- Xcode Command Line Tools

### Build and run

```sh
./build.sh
open multiroblox.app
```

### Run all tests

```sh
go test ./...
```

Some integration tests under `internal/robloxapp/` require
`/Applications/Roblox.app` to be installed. They are skipped in CI.

### Project layout

| Path | Purpose |
|---|---|
| `main.go` | App entry point, deeplink dispatch loop, shutdown handler |
| `internal/deeplink/` | kAEGetURL Apple Event bridge (Obj-C) |
| `internal/infoplist/` | Info.plist editor (flips `LSMultipleInstancesProhibited`) |
| `internal/macosapp/` | Cocoa window and event loop (Obj-C) |
| `internal/robloxapp/` | Roblox.app cloning, close-all, open |
| `internal/syncbreaker/` | `sem_unlink("/RobloxPlayerUniq")` wrapper |
| `internal/urlhandler/` | Launch Services / NSWorkspace wrapper |
| `pkg/fspath/` | Path and TMPDir helpers |
| `pkg/ps/` | Cross-platform process table reader |
| `multiroblox.app/` | The shipped .app bundle |
| `build.sh` | Native and universal build script |
| `scripts/build-dmg.sh` | Builds a distributable DMG |

### Architectural guidelines

- **No new heavy dependencies.** The whole point of this app is its small
  footprint. Every Go module is an ongoing maintenance burden.
- **No darwinkit, no CocoaPods, no Swift.** The Cocoa bridge is intentionally
  tiny. Anything that needs a UI should extend `internal/macosapp/app.m`
  with a small Obj-C class and call back into Go via cgo `//export`.
- **Keep cgo exports to a minimum.** Each `//export` function creates a new
  C symbol that must be kept in sync with the .h file. Prefer a single
  bridge per concern (deeplink delivery, window lifecycle) over many
  one-off bridges.
- **Test the public API, not the internals.** Unit tests should exercise
  the exported functions of each package. The Objective-C side is not
  unit-testable from Go; rely on manual smoke tests for that.

## Release process

Maintainers only. The flow is:

1. Update `CHANGELOG.md`. Move the `[Unreleased]` section to a new
   versioned section.
2. Bump the version in `multiroblox.app/Contents/Info.plist`
   (`CFBundleShortVersionString` and `CFBundleVersion`).
3. Tag the commit: `git tag -s v2.x.y`.
4. Push the tag: `git push origin v2.x.y`.
5. The release workflow builds a universal binary, signs and notarizes
   the .app, produces a DMG, and attaches everything to a GitHub Release.

## License

By contributing, you agree that your contributions will be licensed under
the MIT License. See [LICENSE](LICENSE).

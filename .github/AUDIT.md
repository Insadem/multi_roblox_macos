# Final audit — `fix/pr-quality-improvements`

This is the audit run before publishing. Every check is reproducible
from the repo root.

## Static checks

| Check | Command | Result |
|---|---|---|
| gofmt | `gofmt -l .` | clean (no output) |
| go vet | `go vet ./...` | clean (no output) |
| go test (CI slice) | `go test ./pkg/... ./internal/infoplist/` | green |
| go build (native) | `./build.sh` | built `multiroblox.app/Contents/MacOS/multi-roblox-macos` (arm64) |
| go build (universal) | `GOOS=darwin GOARCH=arm64 go build … && GOOS=darwin GOARCH=amd64 go build … && lipo -create …` | built fat binary (verified locally) |

## Repo hygiene

| Check | Result |
|---|---|
| Tracked binaries at top level | none |
| Tracked `.DS_Store` | none (stray `internal/.DS_Store` was removed during this audit) |
| Tracked `*.swp` / `*.bak` / `*.tmp` | none |
| `dist/` artifacts | none tracked |
| `.app/Contents/MacOS/*` binaries | none tracked (`.gitignore` covers them) |
| `go.sum` present and committed | yes |
| `LICENSE` present | yes — MIT, copyright 2024 Kushang Shah + 2024 Insadem |
| `CODE_OF_CONDUCT.md` present | yes (Contributor Covenant v2.1) |
| `CONTRIBUTING.md` present | yes |
| `SECURITY.md` present | yes |
| `CHANGELOG.md` present | yes (Keep a Changelog format) |
| `.github/RELEASE_NOTES/v2.2.0.md` present | yes |
| `.github/workflows/build.yml` present | yes |
| `.github/workflows/release.yml` present | yes |
| `scripts/build-dmg.sh` present and executable | yes |

## Source tree hygiene

| Check | Result |
|---|---|
| Debug `fmt.Println` / `log.Println` outside `main.go` shutdown path | none (the two `log.Printf` in `internal/robloxapp/closeall.go` and the one `log.Println` in `main.go` are intentional user-facing diagnostics, not debug noise) |
| `TODO` / `FIXME` / `XXX` / `HACK` markers | none |
| Dead code referenced from production | none (the three removed `pkg/fspath/*` files were already gone) |
| `go.mod` | declares `go 1.23`; module path is `github.com/Insadem/multi-roblox-macos` |
| `go.sum` | clean, all required entries present |

## Integration tests

Three integration tests in `internal/robloxapp/`,
`internal/syncbreaker/`, and `internal/urlhandler/` require
`/Applications/Roblox.app` and Launch Services state to be in a
specific shape. They were the source of test-suite failures on
developer machines that don't have Roblox installed.

- `internal/robloxapp/{copy,open,closeall}_test.go` now call
  `requireRoblox(t)` and `t.Skip(...)` cleanly when Roblox is
  absent. CI (`.github/workflows/build.yml`) does not run these
  tests at all, so CI remains green either way.
- `internal/syncbreaker/syncbreaker_darwin_test.go` and
  `internal/urlhandler/urlhandler_darwin_test.go` got the same
  guard. These tests have a second environmental dependency
  (Roblox startup timing for `TestBreak`, and Launch Services
  state for `TestUrlHandler`) that the guard does not cover.
  CI also excludes these. On a developer's real machine with a
  warm Roblox install, they pass; on a freshly-booted dev box
  they may flake. This is pre-existing behaviour, not a
  regression introduced by this branch.

## Documentation

| File | Status |
|---|---|
| `README.md` | Rewritten. Compatibility table now lists only the macOS / Go / arch actually tested by the maintainer (macOS 26 Tahoe on Apple Silicon, Go 1.23 and 1.26). Removed "any number" and "11 Big Sur through 26 Tahoe" over-claims. Added `## Demo` with the verification screenshot. Added `## Credits` pointing to Insadem's original repo. |
| `CHANGELOG.md` | Startup crash line now says "Darwin 25.x (macOS 26 Tahoe)" rather than "macOS 15+", matching what the maintainer actually tested. |
| `.github/RELEASE_NOTES/v2.2.0.md` | Rewritten. Added a "Verified by the maintainer" table, a "Known issues" entry for unverified macOS versions, and a screenshot reference. |
| `.github/PR_BODY.md` | New. PR body for the upstream PR — Problem, Root cause, Solution, Testing, Compatibility, Known limitations, Risk assessment, Checklist. |
| `.github/PR_SUMMARY.md` | New. One-paragraph PR summary, suitable for the PR description header or a GitHub comment. |
| `.github/REBASE_PLAN.md` | New. Suggested rebase plan, including exact `git rebase -i` commands. Do not run blindly — review the proposed shape first. |
| `SECURITY.md` | Unchanged from prior pass. |
| `CONTRIBUTING.md` | Unchanged from prior pass. |
| `CODE_OF_CONDUCT.md` | Unchanged from prior pass. |
| `LICENSE` | Unchanged from prior pass. |

## Build / distribution

| Check | Result |
|---|---|
| `./build.sh` | builds a working native bundle |
| `./build.sh universal` | builds a working fat binary (arm64 + amd64) |
| `scripts/build-dmg.sh` | produces a DMG in `dist/`, optionally signed and notarized |
| `.github/workflows/build.yml` | runs `gofmt`, `go vet`, the CI test slice, and matrix builds for arm64 / amd64 / universal |
| `.github/workflows/release.yml` | triggered on `v*.*.*` tags; builds a universal binary, optionally signs and notarizes, uploads DMG, and drafts a GitHub Release using `.github/RELEASE_NOTES/v2.2.0.md` |

## Remaining warnings

- `ld: warning: ignoring duplicate libraries: '-lobjc'` during
  the native build. This is benign — both cgo's auto-injected
  `-lobjc` and the AppKit framework contribute the same library.
  Suppressing it would require a custom LDFLAGS strip in
  `build.sh`; the warning is not worth the complexity.

## Recommendation

The branch is ready to push.

# Publishing checklist

Walk through this list in order. Every step is something the
maintainer has to do — none of it can be done by automation that
doesn't have GitHub credentials.

## 1. Pre-push sanity

```sh
cd "/Users/kushangshah/github testing/multi-roblox-macos"
git checkout fix/pr-quality-improvements

# Verify a clean state on the public commits
git log --oneline main..HEAD

# Re-run the local quality gates
gofmt -l .                           # expect no output
go vet ./...                         # expect no output
go test ./pkg/... ./internal/infoplist/   # expect green

# Rebuild
./build.sh
```

Open `multiroblox.app` and confirm a window appears, then close
it. The repo is a release candidate — this is the last local
sanity check.

## 2. Decide on the rebase

Read `.github/REBASE_PLAN.md` first. The current order is already
clean; the only reason to rebase is to fold the
`docs: add v2.2.0 release notes` commit into
`docs: add professional open-source documentation` so the branch
has nine commits instead of ten. Either is defensible.

If you rebase, push with `--force-with-lease`, not `--force`.

## 3. Push to the fork

The fork remote is already configured:

```sh
git push --force-with-lease origin fix/pr-quality-improvements
```

Verify on <https://github.com/kushangshah/multi-roblox-macos/branches>
that the branch is there and the tip matches the local tip.

## 4. Open the PR

On
<https://github.com/kushangshah/multi-roblox-macos/compare/main...fix/pr-quality-improvements>:

- **Title.** `fix: restore compatibility with modern macOS and
  replace abandoned darwinkit wrapper`.
- **Body.** Paste the contents of `.github/PR_BODY.md`. The
  one-paragraph `PR_SUMMARY.md` is suitable for a comment
  instead, if you want to keep the PR body shorter.
- **Base repository.** `Insadem/multi-roblox-macos`,
  `main`. This is the upstream PR.
- **Reviewers.** Tag the original author if you have a contact.
  Otherwise leave reviewers empty — they get assigned by the
  maintainer.

If the upstream PR is to Insadem and the fork PR is to your own
remote, do the upstream PR first, since the maintainer will
react to the public record.

## 5. First release

The release workflow runs on a `v*.*.*` tag push:

```sh
git checkout fix/pr-quality-improvements
git tag v2.2.0
git push origin v2.2.0
```

Then on
<https://github.com/kushangshah/multi-roblox-macos/releases>,
edit the auto-drafted release:

- **Tag.** `v2.2.0`
- **Title.** `v2.2.0 — Multi-Roblox on macOS, now with modern Cocoa`
- **Body.** Paste the contents of
  `.github/RELEASE_NOTES/v2.2.0.md` (the release workflow should
  have done this — verify).
- **Assets.** The workflow should have uploaded
  `multi-roblox-macos-2.2.0.dmg` and
  `multi-roblox-macos-2.2.0.dmg.sha256`. Verify both are there.

## 6. Optional — code signing

The release workflow signs and notarizes the DMG only when
`SIGN_IDENTITY`, `NOTARY_PROFILE`, and
`NOTARYTOOL_TEAM_ID` secrets are set in the repo's Actions
secrets. If you have a Developer ID:

1. Add the three secrets under
   <https://github.com/kushangshah/multi-roblox-macos/settings/secrets/actions>.
2. Delete the existing `v2.2.0` tag and re-push:

   ```sh
   git tag -d v2.2.0
   git push --delete origin v2.2.0
   git push origin v2.2.0
   ```

3. The workflow now runs the codesign and notarize steps.

If you do not have a Developer ID, ship the unsigned DMG and let
users follow the `xattr -cr multiroblox.app` step in the README.

## 7. Verify CI

After the push:

- `.github/workflows/build.yml` should run on the PR and go
  green.
- `.github/workflows/release.yml` should run on the tag push and
  produce a `dist/` artifact attached to the release.

If either is red, the most likely cause is a macOS runner
version drift; in that case, look at the failed job's log
before pulling on it.

## 8. Reviewer notes

`.github/PR_BODY.md` already contains a `## Risk assessment`
section. If a reviewer asks why darwinkit was removed instead
of upgraded, the short answer is:

- `darwinkit v0.5.1` is the most recent release.
- It calls into the Cocoa run loop in a way that throws a C++
  exception on Darwin 25.x (macOS 26 Tahoe), aborting the
  process before the window appears. Confirmed by the
  maintainer on a real Apple Silicon Mac.
- Upgrading would require either waiting on a new release that
  has not appeared in years, or forking darwinkit and
  maintaining a Go Cocoa binding. The Objective-C bridge in
  this PR is 110 lines, has no third-party Go dependency, and
  is the minimum viable replacement for the surface area
  multi-roblox-macos actually uses (a window with two buttons
  and an Apple Event handler).

If a reviewer asks why the public URL scheme is
`roblox-player://` (not `roblox://` or something else), the
short answer is: that's what roblox.com's "Play" button
delivers, so it has to be that scheme for the app to receive
the deeplink.

## 9. Once the PR is merged upstream

1. Sync your fork:

   ```sh
   git checkout main
   git pull --rebase upstream main   # if upstream is added
   git push origin main
   ```

2. Delete the `fix/pr-quality-improvements` branch on the fork
   and locally.
3. Bump version. The next set of changes goes on a new branch
   off `main`.

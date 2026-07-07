# Suggested rebase plan

> **Do not run this yet.** Review the proposed shape first. Adjust the
> commit subjects if you want them, then run the commands verbatim.
> If something looks off, stop after the first step (`git rebase -i`)
> and edit the todo list in your editor before saving.

## Current branch state

`fix/pr-quality-improvements` is the working branch. It currently
sits on top of `main` (`a64dc1c removed .DS_Store`) and contains 11
of our commits. The branch tip is `dd13f6a`.

```
dd13f6a chore: remove tracked 32MB binary
b810b27 chore: ignore dist/ in gitignore
434a88f docs: add v2.2.0 release notes
c29c951 ci: add GitHub Actions for build, test, and release
48b3512 docs: add professional open-source documentation
93bb133 feat: add build script for native and universal builds
36df6f2 fix: replace darwinkit UI with native Cocoa implementation
c3b4bdc fix: rework deeplink dispatch and shutdown
6b180d4 fix: migrate urlhandler off deprecated LSCopyDefaultHandlerForURLScheme
ff2682e fix: harden error handling and clean up dead code
b1f7623 fix: register roblox-player:// URL scheme in Info.plist
```

The history is already conventional-commits-shaped, but it is
linear-flat. Reviewers tend to read the branch in one go, so the
order and grouping matter.

## Goals

1. **Lead with the user-visible fix.** The startup crash and the
   deeplink registration are the headline. A reviewer should not
   have to scan past chore commits to find them.
2. **Group related plumbing into single commits.** The "fix:"
   commits that touch the deeplink dispatch, the Cocoa UI bridge,
   the URL handler migration, and the Info.plist change are all
   parts of the same root-cause story. Squashing them into one
   `fix:` commit is a defensible choice; splitting them out is
   also defensible. The plan below keeps them as separate
   commits but reorders them so the most user-visible one is
   first.
3. **Land `feat: build script` immediately after the fix.** The
   build script is the entry point reviewers will run; it should
   be available before they reach the docs.
4. **Land CI, docs, and chores last.** These are the supporting
   work. They do not change runtime behaviour.
5. **No force-push to `main`.** The rebase is local to
   `fix/pr-quality-improvements` only.

## Proposed commit order (target)

```
1.  fix: register roblox-player:// URL scheme in Info.plist
2.  fix: replace darwinkit UI with native Cocoa implementation
3.  fix: rework deeplink dispatch and shutdown
4.  fix: migrate urlhandler off deprecated LSCopyDefaultHandlerForURLScheme
5.  fix: harden error handling and clean up dead code
6.  feat: add build script for native and universal builds
7.  ci: add GitHub Actions for build, test, and release
8.  docs: add professional open-source documentation
9.  docs: add v2.2.0 release notes
10. chore: ignore dist/ in gitignore
11. chore: remove tracked 32MB binary
```

This is the current order with no squashing — the existing
subjects are already good. If you want a tighter branch, an
alternative is:

```
1.  fix: register roblox-player:// URL scheme in Info.plist
2.  fix: replace darwinkit UI with native Cocoa implementation
3.  fix: rework deeplink dispatch and shutdown
4.  fix: migrate urlhandler off deprecated Launch Services API
5.  fix: harden error handling and clean up dead code
6.  feat: add build script for native and universal builds
7.  ci: add GitHub Actions for build, test, and release
8.  docs: add professional open-source documentation
9.  chore: ignore dist/ in gitignore
10. chore: remove tracked 32MB binary
```

The "docs: add v2.2.0 release notes" commit is folded into the
docs commit in this alternative.

## Commands (run only after review)

Start the interactive rebase against `main`. The commits in
`fix/pr-quality-improvements` are the 11 commits listed at the top
of this file, so the rebase covers `HEAD~11..HEAD`.

```sh
cd "/Users/kushangshah/github testing/multi-roblox-macos"
git checkout fix/pr-quality-improvements
git rebase -i main
```

Your editor will open with the 11 commits in oldest-to-newest
order. The current order already matches the proposed target
order, so the only thing you usually need to do is **save and
quit**. If you want the tighter shape, edit the todo list to
either `pick` (keep order) or `fixup` (squash into the previous
commit) the entries. Conventional Commits `fixup=` is the right
verb to drop a commit's diff into its parent without keeping the
subject.

A typical todo list for the tighter shape would be:

```
pick b1f7623 fix: register roblox-player:// URL scheme in Info.plist
pick 36df6f2 fix: replace darwinkit UI with native Cocoa implementation
pick c3b4bdc fix: rework deeplink dispatch and shutdown
pick 6b180d4 fix: migrate urlhandler off deprecated LSCopyDefaultHandlerForURLScheme
pick ff2682e fix: harden error handling and clean up dead code
pick 93bb133 feat: add build script for native and universal builds
pick c29c951 ci: add GitHub Actions for build, test, and release
pick 48b3512 docs: add professional open-source documentation
fixup 434a88f docs: add v2.2.0 release notes
pick b810b27 chore: ignore dist/ in gitignore
pick dd13f6a chore: remove tracked 32MB binary
```

Save and quit. If there are no conflicts, the rebase finishes in
a few seconds. After it completes:

```sh
# Verify the new history
git log --oneline -15

# Re-run the local quality gates
gofmt -l .
go vet ./...
go test ./...
./build.sh
open multiroblox.app
```

If `go test ./...` is green and the app still launches and
spawns two Roblox instances, the rebase is safe. Force-push the
rewritten branch:

```sh
git push --force-with-lease origin fix/pr-quality-improvements
```

`--force-with-lease` is safer than `--force` because it refuses
to clobber a remote ref that has moved on since you last fetched.

## If something goes wrong

```sh
# Abort the rebase and start over
git rebase --abort

# Or open a new branch from the pre-rebase tip and start over
git branch backup/fix-pr-quality-improvements-pre-rebase dd13f6a
```

A backup branch is cheap insurance. The current tip is
`dd13f6a`; the command above creates `backup/...` pointing at the
same commit so you can always `git reset --hard
backup/fix-pr-quality-improvements-pre-rebase` to recover.

## What to skip

- **Do not rebase `main`.** `main` is the public branch
  (`a64dc1c removed .DS_Store`). The rebase only touches
  `fix/pr-quality-improvements`.
- **Do not rewrite commit subjects to add a scope.** Subjects
  like `fix(robloxapp): ...` are fine but not required.
  Conventional Commits' `type(scope): subject` form is allowed
  with or without the scope. The current subjects are
  consistent and clear; leave them alone unless you have a
  specific reason to change them.
- **Do not split the `fix:` commits further.** Each one already
  has a single coherent theme. Going finer (e.g. one commit
  per file) makes the history harder to read.
- **Do not reorder past `main` into the new branch.** Anything
  older than `a64dc1c` belongs to Insadem's history and stays
  where it is.

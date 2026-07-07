# multi-roblox-macos

Run multiple [Roblox](https://www.roblox.com) players at the same time on macOS — one per browser account, with teleport between places still working.

## Overview

`multi-roblox-macos` is a small native macOS launcher that lets you play several Roblox games in parallel from the same Mac, each signed in to a different Roblox account. It is useful when you want to test your own game with multiple accounts, play with friends using alts, or run a private server alongside your main account.

Roblox's own `RobloxPlayer.app` declares `LSMultipleInstancesProhibited = true`, which makes Launch Services refuse to start a second copy. This app works around that by:

1. Registering itself as the OS handler for the `roblox-player://` URL scheme.
2. On every deeplink handed to it by the browser, cloning `/Applications/Roblox.app` to a fresh temp directory, flipping `LSMultipleInstancesProhibited` off in the copy's `Info.plist`, and opening it.
3. Cleaning up the temp copy on quit.

The bundled binary is a small, self-contained Go program that links directly against AppKit. There is no Xcode project, no CocoaPods, and no Swift. The whole toolchain is `go build` plus `build.sh`.

## Features

- **Multiple Roblox instances** — play any number of Roblox games in parallel, one per browser account.
- **Apple Silicon support** — builds and runs natively on arm64.
- **Intel support** — builds and runs on x86_64.
- **Universal binary** — `build.sh universal` produces a single fat binary with both architectures.
- **Native macOS app** — small Cocoa window with two buttons (Discord link, "close all instances"). No browser, no Electron, no system tray.
- **Automatic Roblox cloning** — every deeplink spawns a fresh clone; an auto-updated Roblox is picked up on the next click.
- **Modern macOS compatibility** — runs on macOS 11 Big Sur through macOS 26 Tahoe.
- **Teleport support** — teleport between places is preserved, because teleport deeplinks are routed through the same `roblox-player://` handler.

## Installation

### Requirements

- macOS 11 (Big Sur) or newer, on Apple Silicon or Intel
- [Go 1.23+](https://go.dev/dl/) (only required to build from source)
- Xcode Command Line Tools (`xcode-select --install`)

### Clone

```sh
git clone https://github.com/kushangshah/multi-roblox-macos.git
cd multi-roblox-macos
```

### Build

```sh
# Build for the architecture you are running on
./build.sh

# Or build a universal (arm64 + amd64) bundle
./build.sh universal
```

The script writes the binary into `multiroblox.app/Contents/MacOS/multi-roblox-macos`.

### Run

```sh
open multiroblox.app
```

A small window titled "multi-roblox-macos" appears. Leave it running in the background.

To start a Roblox game, go to [roblox.com](https://www.roblox.com) in your browser, sign in to the account you want to play on, and click any "Play" button. The browser hands the deeplink to this app, which spawns a new instance.

To start a *second* instance, switch to a different browser account (or use a different browser profile) and click Play again. Repeat as many times as you like.

To stop everything, click the **close all instances** button in the window.

### Uninstall

1. Quit the multi-roblox-macos window.
2. Drag `multiroblox.app` to the Trash.
3. (Optional) Remove any leftover temp directories:

   ```sh
   rm -rf "$TMPDIR"/multi-roblox-*
   ```

   These are normally cleaned up automatically on quit, but a hard kill of the app can leave a few behind.

The app does not install any daemons, login items, or kernel extensions.

## Usage

```sh
./build.sh           # build for current arch
open multiroblox.app # run the app
```

Then:

1. Click **Play** on roblox.com under account A. The first Roblox instance starts.
2. Switch to account B in the same browser (or another browser profile).
3. Click **Play** again. A second, independent instance starts.
4. Use teleport as normal — it still works because both instances are listening for `roblox-player://` deeplinks.

To terminate all running Roblox processes, click **close all instances** in the multi-roblox-macos window.

## How it works

When you click "Play" on roblox.com, the browser asks the OS to open a `roblox-player://...` URL. By default, macOS routes that URL to RobloxPlayer.app, which refuses to start a second copy.

This app registers itself as the default handler for the `roblox-player://` scheme. macOS now hands every deeplink to multi-roblox-macos instead.

For each deeplink:

1. A unique subdirectory is created in `$TMPDIR` via `os.MkdirTemp`.
2. `/Applications/Roblox.app` is recursively copied (`cp -a`) into that directory.
3. The named semaphore `/RobloxPlayerUniq` — which Roblox uses to detect an already-running player — is destroyed via `sem_unlink`.
4. The copy is opened with `open -a`.
5. After launch, the copy's `Info.plist` is patched to set `LSMultipleInstancesProhibited = false`, so that the next deeplink is not blocked.
6. On quit, every copy is removed (`os.RemoveAll`).

The `kAEGetURL` Apple Event that delivers the deeplink is installed via a small Objective-C bridge in `internal/deeplink/`. The Cocoa UI window is installed via a small Objective-C bridge in `internal/macosapp/`.

## Compatibility

| Component | Versions tested |
|---|---|
| macOS | 11 Big Sur, 12 Monterey, 13 Ventura, 14 Sonoma, 15 Sequoia, 26 Tahoe |
| Architecture | Apple Silicon (arm64), Intel (x86_64), Universal |
| Go | 1.23, 1.24, 1.25, 1.26 |
| Roblox | Any version that ships a `RobloxPlayer.app` under `/Applications` |

## Troubleshooting

### Window doesn't appear

- Confirm the binary exists: `ls -la multiroblox.app/Contents/MacOS/multi-roblox-macos`.
- Try launching the binary directly from a terminal to see error output:
  ```sh
  ./multiroblox.app/Contents/MacOS/multi-roblox-macos
  ```
- Make sure no older copy of the app is still running. Use `Activity Monitor` and search for `multi-roblox-macos`.

### Roblox doesn't launch

- Confirm Roblox is installed: `ls /Applications/Roblox.app`. If it is not, install Roblox from [roblox.com](https://www.roblox.com).
- macOS only allows one app at a time to own a URL scheme. If Roblox itself owns `roblox-player://`, this app re-registers itself on every deeplink. If launching still fails, quit and reopen the multi-roblox-macos window.

### Only one Roblox instance launches

- Some browser extensions intercept `roblox-player://` URLs and route them to the original Roblox app. Disable extensions on roblox.com, or try a different browser.
- If a previous Roblox process is hung, click **close all instances** to terminate it, then try again.

### Permissions issues

- This app does not require Full Disk Access, Accessibility, or any other macOS privacy permission. It only writes to `$TMPDIR` and to copies of `Roblox.app` it owns.

### Gatekeeper / "unidentified developer"

- On the first run, macOS may refuse to open the app because it is not notarized. Right-click (or Control-click) the app, choose **Open**, then click **Open** in the dialog.
- For a permanent fix without right-clicking, strip the quarantine attribute:
  ```sh
  xattr -cr multiroblox.app
  ```

### Code signing

- This repository does not ship signed binaries. If you want to distribute a build that survives Gatekeeper without the `xattr` workaround, sign it with your own Developer ID:
  ```sh
  codesign --force --deep --options runtime \
    --sign "Developer ID Application: Your Name (TEAMID)" \
    multiroblox.app
  ```
- See [`.github/workflows/release.yml`](.github/workflows/release.yml) for a complete release workflow.

## Developer

### Project structure

```
.
├── main.go                          # entry point, deeplink dispatch loop
├── build.sh                         # native and universal build script
├── go.mod / go.sum                  # Go module manifest
├── multiroblox.app/                 # the .app bundle
│   └── Contents/
│       ├── Info.plist               # registers the roblox-player:// scheme
│       ├── MacOS/                   # the compiled binary lives here
│       └── Resources/AppIcon.icns   # app icon
├── internal/
│   ├── deeplink/                    # kAEGetURL Apple Event bridge (Obj-C)
│   ├── infoplist/                   # Info.plist editor (LSMultipleInstancesProhibited)
│   ├── macosapp/                    # Cocoa window + run loop (Obj-C)
│   ├── robloxapp/                   # Roblox.app cloning, close-all, open
│   ├── syncbreaker/                 # sem_unlink("/RobloxPlayerUniq")
│   └── urlhandler/                  # Launch Services / NSWorkspace wrapper
└── pkg/
    ├── fspath/                      # Path + TMPDir helpers
    └── ps/                          # cross-platform process table reader
```

### Building

- `./build.sh` — native build for the host arch
- `./build.sh universal` — fat binary with arm64 and amd64

### Running tests

```sh
go test ./...
```

Unit tests live next to the code they cover (`*_test.go`). Integration tests under `internal/robloxapp/` require `/Applications/Roblox.app` to be installed; they are skipped in CI.

### Building a DMG

The release workflow at `.github/workflows/release.yml` produces a signed DMG via `create-dmg`. To build a DMG locally:

```sh
brew install create-dmg
./scripts/build-dmg.sh
```

The output is `dist/multi-roblox-macos-<version>.dmg`.

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This project is released under the MIT License. See [LICENSE](LICENSE).

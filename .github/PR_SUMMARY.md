# v2.2.0 — One-paragraph summary

Restores `multi-roblox-macos` to a working state on modern macOS by
replacing the abandoned `darwinkit` Cocoa wrapper with a small
Objective-C bridge, rewriting `Info.plist` to declare the
`roblox-player://` URL scheme plus the standard bundle keys, fixing
several silent error-swallowing paths, and hardening the deeplink
dispatch loop and `Close All` shutdown path. Verified by the
maintainer on a real Apple Silicon Mac running macOS 26 Tahoe
(Darwin 25.2.0) with Go 1.23 and 1.26: two Roblox players run
simultaneously, teleport still works, and the temp clones are
cleaned up on quit.

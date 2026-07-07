# Security Policy

## Supported versions

| Version | Supported |
|---|---|
| 2.2.x   | ✅ |
| 2.1.x   | ❌ |
| 2.0.x   | ❌ |
| < 2.0   | ❌ |

## Reporting a vulnerability

If you discover a security issue, please report it privately:

- **Email:** open a private security advisory at
  <https://github.com/kushangshah/multi-roblox-macos/security/advisories/new>.

Please **do not** open a public issue for security problems. Give us a
reasonable amount of time to investigate and patch before any public
disclosure.

## What to include

- A clear description of the vulnerability and its impact.
- Steps to reproduce, or a proof-of-concept.
- The exact version affected (commit SHA if possible).
- Your environment (macOS version, Go version if built from source).

## Response timeline

- **Initial acknowledgement:** within 7 days.
- **Triage and severity assessment:** within 14 days.
- **Patch for critical issues:** as soon as practical, typically within 30 days.

We will keep you informed of progress and credit you in the fix commit
unless you prefer to remain anonymous.

## Threat model

This app:

- Reads the system process table to find running RobloxPlayer processes.
- Reads and writes `Info.plist` files inside `$TMPDIR` (clones of
  `/Applications/Roblox.app`).
- Calls Launch Services to register itself as the default handler for
  `roblox-player://` URLs.
- Opens a Cocoa window and runs the main run loop.

It does not:

- Make any network requests.
- Read or write files outside `$TMPDIR` (other than the system's URL
  scheme registration).
- Run with elevated privileges.
- Load third-party code at runtime.
- Persist any data on disk.

The named-semaphore trick (`sem_unlink("/RobloxPlayerUniq")`) is a
per-process kernel object; it has no security implications beyond
allowing the second Roblox instance to start.

## Out of scope

- Vulnerabilities in `/Applications/Roblox.app` itself. Report those to
  Roblox.
- Vulnerabilities in macOS or AppKit. Report those to Apple.
- Vulnerabilities in Go's cgo or runtime. Report those to the Go team.

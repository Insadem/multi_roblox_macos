// Package macosapp is a minimal Cocoa UI for multi-roblox-macos.
//
// We avoid the darwinkit dependency because darwinkit (v0.5.1, July 2024)
// is incompatible with Go 1.21+ and modern macOS Cocoa runtimes — its
// wrapper around [NSApplication run] throws a C++ exception on startup
// when launched inside a .app bundle on macOS 15+.
//
// Instead, this package bridges to a tiny Obj-C implementation in
// app.m. The Go side just installs callbacks and blocks on the run
// loop.
package macosapp

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation -framework AppKit
#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#import "app.h"
*/
import "C"

import (
	"github.com/Insadem/multi-roblox-macos/internal/robloxapp"
)

// quitHook is installed by Run; it runs on the main thread when the user
// closes the window.
var quitHook func()

// onCloseAllClicked is the C callback target for the "close all" button.
//
//export GoOnCloseAllClicked
func GoOnCloseAllClicked() {
	robloxapp.CloseAll()
}

// onQuit is the C callback target for window close. We invoke the user
// cleanup and then ask AppKit to terminate the application.
//
//export GoOnQuit
func GoOnQuit() {
	if quitHook != nil {
		quitHook()
	}
	C.terminateApp()
}

// Run blocks on the Cocoa event loop. The clean callback runs on the
// main thread when the user closes the window.
func Run(clean func()) {
	quitHook = clean
	C.runApp()
}

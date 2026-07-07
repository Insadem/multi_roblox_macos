// Package urlhandler reads and writes the system's default handler for URL
// schemes (e.g. roblox-player://) using the Launch Services framework.
//
// We avoid LSCopyDefaultHandlerForURLScheme, which has been deprecated
// since macOS 10.15 in favour of NSWorkspace's
// URLForApplicationToOpenURL:. The C bridge is kept in a .m file so we
// can use the proper Foundation/NSWorkspace APIs and ARC-clean memory
// management.
package urlhandler

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation -framework AppKit
#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#import <stdlib.h>

static int setHandler(const char *bundleIdentifier, const char *urlScheme) {
    NSString *scheme = [NSString stringWithUTF8String:urlScheme];
    NSString *bundleId = [NSString stringWithUTF8String:bundleIdentifier];

    NSURL *appURL = [[NSWorkspace sharedWorkspace]
        URLForApplicationWithBundleIdentifier:bundleId];
    if (appURL == nil) {
        // The bundle is not installed (e.g. user deleted Roblox).
        return 1;
    }

    CFErrorRef err = NULL;
    BOOL ok = LSSetDefaultHandlerForURLScheme(
        (__bridge CFStringRef)scheme,
        (__bridge CFStringRef)bundleId);

    if (err != NULL) {
        CFRelease(err);
    }
    return ok ? 0 : 1;
}

static int checkHandler(const char *bundleIdentifier, const char *urlScheme) {
    NSString *scheme = [NSString stringWithUTF8String:urlScheme];
    NSString *bundleId = [NSString stringWithUTF8String:bundleIdentifier];

    NSURL *appURL = [[NSWorkspace sharedWorkspace]
        URLForApplicationToOpenURL:[NSURL URLWithString:scheme]];
    if (appURL == nil) {
        return 1;
    }

    NSString *defaultBundleId = [[NSBundle bundleWithURL:appURL] bundleIdentifier];
    if (defaultBundleId == nil) {
        return 1;
    }

    return [defaultBundleId isEqualToString:bundleId] ? 0 : 1;
}
*/
import "C"
import (
	"unsafe"
)

const (
	// ROBLOX_BUNDLE_IDENTIFIER is the bundle id of RobloxPlayer.app. It is
	// the only id we ever check against, and the id we restore when this
	// app quits.
	ROBLOX_BUNDLE_IDENTIFIER = "com.roblox.RobloxPlayer"
)

// Set registers bundleIdentifier as the default OS handler for urlScheme.
// Returns true on success.
func Set(bundleIdentifier, urlScheme string) bool {
	cBundle := C.CString(bundleIdentifier)
	defer C.free(unsafe.Pointer(cBundle))

	cURL := C.CString(urlScheme)
	defer C.free(unsafe.Pointer(cURL))

	return C.setHandler(cBundle, cURL) == 0
}

// Check reports whether bundleIdentifier is currently the default OS
// handler for urlScheme.
func Check(bundleIdentifier, urlScheme string) bool {
	cBundle := C.CString(bundleIdentifier)
	defer C.free(unsafe.Pointer(cBundle))

	cURL := C.CString(urlScheme)
	defer C.free(unsafe.Pointer(cURL))

	return C.checkHandler(cBundle, cURL) == 0
}

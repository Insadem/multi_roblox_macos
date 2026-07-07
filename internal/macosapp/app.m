// app.m — minimal Cocoa UI for multi-roblox-macos.
//
// We do all of the window / event-loop setup here in pure Objective-C
// so the Go side never has to touch the Obj-C runtime directly.
// The bridge functions runApp() and terminateApp() are called from
// Go via cgo.

#import "app.h"

// WindowDelegate handles the user closing the window. We forward the
// event back to Go so it can do its cleanup, then terminate the app.
@interface MRWindowDelegate : NSObject <NSWindowDelegate>
@end

@implementation MRWindowDelegate
- (void)windowWillClose:(NSNotification *)notification {
    GoOnQuit();
    // The window is going away; ask NSApp to exit so the run loop ends.
    [NSApp terminate:nil];
}
@end

// MRActionBridge holds the target/action wiring for our buttons. We
// need a real Objective-C object to be the target of the button's
// action selector, so we use this.
@interface MRActionBridge : NSObject
- (void)onCloseAll:(id)sender;
- (void)onDiscordClicked:(id)sender;
@end

@implementation MRActionBridge
- (void)onCloseAll:(id)sender {
    GoOnCloseAllClicked();
}
- (void)onDiscordClicked:(id)sender {
    // /usr/bin/open on a URL is the simplest cross-version way to open
    // it in the user's default browser. We could use
    // [NSWorkspace openURL:] directly, but launching a process keeps
    // the action re-entrancy obvious.
    [NSTask launchedTaskWithLaunchPath:@"/usr/bin/open"
                              arguments:@[@"https://discord.gg/AwNAa7utbY"]];
}
@end

// runApp is the single C entry point called from Go. It blocks on
// [NSApp run] until the user quits.
void runApp(void) {
    @autoreleasepool {
        [NSApplication sharedApplication];
        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

        NSRect frame = NSMakeRect(0, 0, 320, 160);
        NSUInteger style = NSWindowStyleMaskTitled
                          | NSWindowStyleMaskClosable
                          | NSWindowStyleMaskMiniaturizable;
        NSWindow *win = [[NSWindow alloc] initWithContentRect:frame
                                                    styleMask:style
                                                      backing:NSBackingStoreBuffered
                                                        defer:NO];
        [win setTitle:@"multi-roblox-macos"];
        [win center];

        MRWindowDelegate *winDel = [[MRWindowDelegate alloc] init];
        [win setDelegate:winDel];

        NSStackView *stack = [[NSStackView alloc] initWithFrame:frame];
        [stack setOrientation:NSUserInterfaceLayoutOrientationVertical];
        [stack setDistribution:NSStackViewDistributionFillEqually];
        [stack setAlignment:NSLayoutAttributeCenterX];
        [stack setSpacing:10];

        NSTextField *label = [NSTextField labelWithString:@"start any roblox game via browser"];
        [label setAlignment:NSTextAlignmentCenter];

        MRActionBridge *bridge = [[MRActionBridge alloc] init];

        NSButton *discordBtn = [NSButton buttonWithTitle:@"join discord server"
                                                  target:bridge
                                                  action:@selector(onDiscordClicked:)];
        NSButton *closeBtn = [NSButton buttonWithTitle:@"close all instances"
                                                target:bridge
                                                action:@selector(onCloseAll:)];

        [stack addArrangedSubview:label];
        [stack addArrangedSubview:discordBtn];
        [stack addArrangedSubview:closeBtn];

        [win.contentView addSubview:stack];
        stack.translatesAutoresizingMaskIntoConstraints = NO;
        [NSLayoutConstraint activateConstraints:@[
            [stack.leadingAnchor constraintEqualToAnchor:win.contentView.leadingAnchor constant:20],
            [stack.trailingAnchor constraintEqualToAnchor:win.contentView.trailingAnchor constant:-20],
            [stack.topAnchor constraintEqualToAnchor:win.contentView.topAnchor constant:10],
            [stack.bottomAnchor constraintEqualToAnchor:win.contentView.bottomAnchor constant:-10],
        ]];

        [win makeKeyAndOrderFront:nil];
        [NSApp activateIgnoringOtherApps:YES];
        [NSApp run];
    }
}

// terminateApp is called from Go to ask NSApp to exit. We dispatch
// asynchronously so the caller (which is the window's main-thread
// callback) can unwind first.
void terminateApp(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        [NSApp terminate:nil];
    });
}

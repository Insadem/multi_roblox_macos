#ifndef macosapp_app_h
#define macosapp_app_h

#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>

// Go-side callbacks. These are exported by app.go via cgo //export.
extern void GoOnCloseAllClicked(void);
extern void GoOnQuit(void);

// C-side entry points implemented in app.m.
void runApp(void);
void terminateApp(void);

#endif /* macosapp_app_h */

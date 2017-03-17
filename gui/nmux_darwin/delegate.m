#import "delegate.h"

@implementation AppDelegate

- (void)applicationMenuSelected:(NSMenuItem *)menu {
  appMenuSelected((char *)[[menu title] UTF8String]);
}

- (void)applicationWillUpdate:(NSNotification *)notification {
  // ???
}

- (void)applicationDidFinishLaunching:(NSNotification *)notification {
  NSLog(@"Launching");
  [[NSRunningApplication currentApplication]
   activateWithOptions:NSApplicationActivateAllWindows|NSApplicationActivateIgnoringOtherApps];

#ifndef NMUX_CGO
  NmuxScreen *view = (NmuxScreen *)newWindow(0, 0);
  spam(view);
#endif

  appStarted();
}

- (void)applicationWillTerminate:(NSNotification *)notification {
  NSLog(@"Terminate");
  appStopped();
}

- (void)applicationWillHide:(NSNotification *)notification {
  NSLog(@"Hiding");
  appHidden();
}

@end

/* vim: set ft=objc ts=2 sw=2 et :*/

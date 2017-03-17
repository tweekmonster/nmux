#ifndef DELEGATE_H
#define DELEGATE_H
#import <Cocoa/Cocoa.h>
#import "cgo_extern.h"
#import "screen.h"
#import "bridge.h"

@interface AppDelegate : NSObject <NSApplicationDelegate>
- (void)applicationMenuSelected:(NSMenuItem *)menu;
@end
#endif /* ifndef DELEGATE_H */

/* vim: set ft=objc ts=2 sw=2 et :*/

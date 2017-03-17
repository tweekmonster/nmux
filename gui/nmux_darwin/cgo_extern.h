#import <Cocoa/Cocoa.h>

#ifndef NMUX_CGO

#define appStarted() NSLog(@"appStarted")
#define appStopped() NSLog(@"appStopped")
#define appHidden() NSLog(@"appHidden")
#define inputEvent(win, key) NSLog(@"inputEvent: %lu - %s", win, key)
#define winMoved(win, x, y) NSLog(@"winMoved: %lu - Pos: %dx%d", win, x, y)
#define winResized(win, w, h, gw, gh) NSLog(@"winResized: %lu - Size: %dx%d, Grid: %dx%d", win, w, h, gw, gh)
#define winClosed(win) NSLog(@"winClosed: %lu", win)
#define winFocused(win) NSLog(@"winFocused: %lu", win)
#define winFocusLost(win) NSLog(@"winFocusLost: %lu", win)
#define appMenuSelected(title) NSLog(@"appMenuSelected: %s", title)
#define windowMenuSelected(win, title) NSLog(@"windowMenuSelected: %lu, '%s'", win, title)

#else

extern void appStarted(void);
extern void appStopped(void);
extern void appHidden(void);
extern void inputEvent(uintptr_t, char *);
extern void winMoved(uintptr_t, int, int);
extern void winResized(uintptr_t, int, int, int, int);
extern void winClosed(uintptr_t);
extern void winFocused(uintptr_t);
extern void winFocusLost(uintptr_t);
extern void appMenuSelected(char *);
extern void windowMenuSelected(uintptr_t, char *);

#endif

/* vim: set ft=objc ts=2 sw=2 et :*/

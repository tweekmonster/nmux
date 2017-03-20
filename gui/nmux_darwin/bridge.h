#import <Cocoa/Cocoa.h>

#ifndef BRIDGE_H
#define BRIDGE_H
#import "nmux.h"
#import "delegate.h"
#import "screen.h"
#import "misc.h"

void startApp(void);
void stopApp(void);
uintptr_t newWindow(int, int);
void setGridSize(uintptr_t, int, int);
void drawText(uintptr_t, const char *, int, int, uint8_t, int32_t, int32_t, int32_t);
void drawRepeatedText(uintptr_t, unichar, int, int, uint8_t, int32_t, int32_t, int32_t);
void clearScreen(uintptr_t, int32_t);
void scrollScreen(uintptr_t, int, int, int, int, int, int32_t);
void flush(uintptr_t, int, int, int, const char *, int, uint8_t, int32_t, int32_t, int32_t);
void getCellSize(int*, int*);
#endif /* ifndef BRIDGE_H */

#ifndef NMUX_CGO

// Don't use GCD when running standalone.
#define DISPATCH_S(block) (block)()
#define DISPATCH_A(block) (block)()
#define DISPATCH_D(block, delay) (block)()

void spam(NmuxScreen *);

#else

#define DISPATCH_S(block) dispatch_sync(dispatch_get_main_queue(), (block))
#define DISPATCH_A(block) dispatch_async(dispatch_get_main_queue(), (block))
#define DISPATCH_D(block, delay) dispatch_after(dispatch_time(DISPATCH_TIME_NOW, delay * NSEC_PER_MSEC), dispatch_get_main_queue(), (block))

#endif

/* vim: set ft=objc ts=2 sw=2 et : */

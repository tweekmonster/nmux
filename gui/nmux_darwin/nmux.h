#ifndef NMUX_H
#define NMUX_H
#import <Cocoa/Cocoa.h>
#import "misc.h"

@interface nmux : NSObject
+ (void)setup;

+ (void)setFontFamily:(NSString *)name size:(CGFloat)size;
+ (NSFont *)font;
+ (CGFloat)descent;
+ (CGFloat)firstCharPos;
+ (NSSize)cellSize;
+ (NSSize)minGridSize;
+ (NSSize)fitGrid:(NSSize)size;

+ (void)setLastWinFrame:(NSRect)frame;
+ (NSRect)lastWinFrame;
@end
#endif /* ifndef NMUX_H */

/* vim: set ft=objc ts=2 sw=2 et :*/

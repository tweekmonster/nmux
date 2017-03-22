#ifndef SCREEN_H
#define SCREEN_H
#import <Cocoa/Cocoa.h>
#import <Carbon/Carbon.h>
#import "ops.h"

@interface NmuxScreen : NSView <NSWindowDelegate>
{
  NSSize _grid;
  Mode _state;

  CGLayerRef screenLayer;
  CGLayerRef cursorLayer;
  NSRect cursorRect;
  BOOL cursorUpdate;
  BOOL didFlush;

  NSPoint lastMouseCoords;
  CGAffineTransform transform;
  CGAffineTransform cursorTransform;
  NSMutableArray *flushOps;
  NSLock *drawLock;

  CGColorRef viewBg;

  unichar *runChars;
  CGGlyph *runGlyphs;
  CGPoint *runPositions;
  size_t runMaxLength;
}

@property (atomic) NSSize grid;
@property (atomic) Mode state;

- (void)setGridSize:(NSSize)size;
- (void)addDrawOp:(DrawOp *)op;
- (void)flushDrawOps:(const char *)character charWidth:(int)width
                 pos:(NSPoint)cursorPos attrs:(TextAttr)attrs;
@end
#endif /* ifndef SCREEN_H */

/* vim: set ft=objc ts=2 sw=2 et :*/

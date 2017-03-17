#import "nmux.h"

static NSFont *font;
static NSSize cellSize;
static NSSize minGridSize = {80, 20};
static NSRect lastWinFrame;

@implementation nmux

+ (void)setup {
  [self setFontFamily:nil size:0];
}

+ (void)setFontFamily:(NSString *)name size:(CGFloat)size {
  NSFont *f;

  if (name == nil && font != nil) {
    name = [font familyName];
  }

  if (name == nil) {
    f = [NSFont userFixedPitchFontOfSize:size];
  } else {
    f = [[NSFontManager sharedFontManager]
         fontWithFamily:name
                 traits:NSUnboldFontMask|NSUnitalicFontMask
                 weight:5
                   size:size];
  }

  cellSize.width = ceil([f maximumAdvancement].width);
  cellSize.height = ceil([f ascender] + ABS([f descender]));

  if (font != nil) {
    [font release];
  }
  font = [f retain];
}

+ (NSFont *)font {
  return font;
}

+ (CGFloat)descent {
  return ABS([font descender]);
}

+ (CGFloat)firstCharPos {
  return (cellSize.width - [font maximumAdvancement].width) / 2;
}

+ (NSSize)cellSize {
  return cellSize;
}

+ (NSSize)minGridSize {
  return minGridSize;
}

+ (void)setLastWinFrame:(NSRect)frame {
  lastWinFrame = frame;
}

+ (NSRect)lastWinFrame {
  if (NSEqualRects(lastWinFrame, NSZeroRect)) {
    lastWinFrame.size = NSSizeMultiply(minGridSize, cellSize);
    NSRect screenRect = [[NSScreen mainScreen] visibleFrame];
    lastWinFrame.origin.x = 0;
    lastWinFrame.origin.y = NSHeight(screenRect) - NSHeight(lastWinFrame);
  }
  return lastWinFrame;
}

+ (NSSize)fitGrid:(NSSize)size {
  size.width -= fmod(size.width, cellSize.width);
  size.height -= fmod(size.height, cellSize.height);
  return size;
}

- (void)dealloc {
  if (font != nil) {
    [font release];
  }

  [super dealloc];
}

@end

/* vim: set ft=objc ts=2 sw=2 et :*/

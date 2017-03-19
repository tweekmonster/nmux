#import "nmux.h"

static NSFont *_currentFont;
static NSMutableArray *_fontCache;
static CGSize _cellSize;
static CGSize _minGridSize = {80, 20};
static NSRect _lastWinFrame;


void nmux_Init() {
  _fontCache = [[NSMutableArray alloc] init];
  nmux_SetFont(nil, 0);
}

CGSize nmux_CellSize() {
  return _cellSize;
}

CGSize nmux_MinGridSize() {
  return _minGridSize;
}

CGSize nmux_FitToGrid(CGSize size) {
  size.width -= fmod(size.width, _cellSize.width);
  size.height -= fmod(size.height, _cellSize.height);
  return size;
}

NSRect nmux_LastWindowFrame() {
  if (NSEqualRects(_lastWinFrame, NSZeroRect)) {
    _lastWinFrame.size = NSSizeMultiply(_minGridSize, _cellSize);
    NSRect screenRect = [[NSScreen mainScreen] visibleFrame];
    _lastWinFrame.origin.x = 0;
    _lastWinFrame.origin.y = NSHeight(screenRect) - NSHeight(_lastWinFrame);
  }
  return _lastWinFrame;
}

void nmux_SetLastWindowFrame(NSRect frame) {
  _lastWinFrame = frame;
}

void nmux_SetFont(NSString *name, CGFloat size) {
  NSFont *f;

  if (name == nil && _currentFont != nil) {
    name = [_currentFont familyName];
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

  _cellSize.width = ceil([f maximumAdvancement].width);
  _cellSize.height = ceil([f ascender] + ABS([f descender]));

  if (_currentFont != nil) {
    [_currentFont release];
  }
  _currentFont = [f retain];
}

NSFont *nmux_CurrentFont() {
  if (_currentFont == nil) {
    nmux_SetFont(nil, 0);
  }
  return _currentFont;
}

CGFloat nmux_FontDescent(NSFont *font) {
  if (font == nil) {
    font = nmux_CurrentFont();
  }

  return ABS([font descender]);
}

CGFloat nmux_InitialCharPos(NSFont *font) {
  if (font == nil) {
    font = nmux_CurrentFont();
  }
  return (_cellSize.width - [font maximumAdvancement].width) / 2;
}

/* vim: set ft=objc ts=2 sw=2 et :*/

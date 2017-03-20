#import "nmux.h"

static CTFontRef _currentFont;
static NSMutableArray *_fontCache;
static CGSize _cellSize;
static CGSize _minGridSize = {80, 20};
static NSRect _lastWinFrame;


void nmux_Init() {
  _currentFont = nil;
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
  CTFontRef font;

  if (name == nil && _currentFont != nil) {
    name = [(NSString *)CTFontCopyFamilyName(_currentFont) autorelease];
  }

  if (name == nil) {
    font = (CTFontRef)[NSFont userFixedPitchFontOfSize:size];
  } else {
    font = CTFontCreateWithName((CFStringRef)name, size, NULL);
  }

  _cellSize.width = ceil([(NSFont *)font maximumAdvancement].width);
  _cellSize.height = ceil(CTFontGetAscent(font) + ABS(CTFontGetDescent(font)));

  if (_currentFont != NULL) {
    CFRelease(_currentFont);
  }
  _currentFont = font;
  CFRetain(font);
  [_fontCache removeAllObjects];
}

CTFontRef nmux_GetFontForChars(const unichar *chars, UniCharCount count, CTFontRef tryFont) {
    CGGlyph glyphs[count];

    if (tryFont == nil) {
      tryFont = _currentFont;
    }

    if (CTFontGetGlyphsForCharacters((CTFontRef)tryFont, chars,
                                     glyphs, count)) {
      return tryFont;
    }

    for (id font in _fontCache) {
      if (CTFontGetGlyphsForCharacters((CTFontRef)font, chars,
                                       glyphs, count)) {
        return (CTFontRef)font;
      }
    }

    CFStringRef str = CFStringCreateWithCharacters(NULL, chars, count);
    CTFontRef newFont = CTFontCreateForString(_currentFont, str,
                                              CFRangeMake(0, count));
    CFRelease(str);

    if (!CTFontGetGlyphsForCharacters((CTFontRef)newFont, chars,
                                      glyphs, count)) {
      CFRelease(newFont);
      return nil;
    }

    if (newFont != NULL) {
      NSLog(@"Caching font: %@", newFont);
      [_fontCache addObject:(id)newFont];
      CFRelease(newFont);
    }

    return newFont;
}

CTFontRef nmux_CurrentFont() {
  if (_currentFont == nil) {
    nmux_SetFont(nil, 0);
  }
  return _currentFont;
}

CGFloat nmux_FontDescent(CTFontRef font) {
  if (font == nil) {
    font = nmux_CurrentFont();
  }

  return ABS(CTFontGetDescent(font));
}

CGFloat nmux_InitialCharPos(CTFontRef font) {
  if (font == nil) {
    font = nmux_CurrentFont();
  }
  return (_cellSize.width - [(NSFont *)font maximumAdvancement].width) / 2;
}

/* vim: set ft=objc ts=2 sw=2 et :*/

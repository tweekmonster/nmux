#include "text_darwin.h"

static NSFont *curFont = nil;
static CGPoint drawPos;
static CGColorSpaceRef colorspace = NULL;

bool set_font(const char *name, CGFloat size, CGFloat *width, CGFloat *ascent,
              CGFloat *descent, CGFloat *leading) {
  NSString *family = [NSString stringWithUTF8String:name];
  NSFont *newFont = [[NSFontManager sharedFontManager] fontWithFamily:family traits:NSUnboldFontMask|NSUnitalicFontMask weight:5 size:size];
  if (newFont != nil) {
    if (curFont != nil) {
      [curFont release];
    }
    CGGlyph g;
    UniChar c = 'X';
    curFont = newFont;

    if (get_glyph(c, &g)) {
      *width = [curFont advancementForGlyph:g].width;
      *ascent = CTFontGetAscent((CTFontRef)curFont);
      *descent = CTFontGetDescent((CTFontRef)curFont);
      *leading = CTFontGetLeading((CTFontRef)curFont);
      drawPos = CGPointMake(0, *descent);
    }

    return true;
  }

  return false;
}

bool get_glyph(UniChar rune, CGGlyph *glyph) {
  CGGlyph glyphs[1];
  UniChar chars[1];
  chars[0] = rune;

  if (CTFontGetGlyphsForCharacters((CTFontRef)curFont, chars, glyphs, 1)) {
    *glyph = glyphs[0];
    return true;
  }

  return false;
}

void draw_glyph(CGContextRef ctx, CGGlyph g, int w, int h, CGColorRef fg, CGColorRef bg) {
  CGRect rect = CGRectMake(0, 0, w, h);

  // Fill background
  CGContextSetFillColorWithColor(ctx, bg);
  CGContextFillRect(ctx, rect);

  // Render Glyph
  CGGlyph glyphs[1];
  glyphs[0] = g;
  CGPoint positions[1];
  positions[0] = drawPos;

  CGContextSetFillColorWithColor(ctx, fg);
  CGContextShowGlyphsAtPositions(ctx, glyphs, positions, 1);

  uint8_t *data = (uint8_t *)CGBitmapContextGetData(ctx);
  int i = 0, l = w*h*4, p;

  while (i < l) {
    p = data[i];
    data[i] = data[i+2];
    data[i + 2] = p;
    i += 4;
  }
}

CGContextRef create_bitmap_context(uint8_t *data, size_t width, size_t height) {
  if (colorspace == NULL) {
    colorspace = CGColorSpaceCreateDeviceRGB();
  }

  // To get sub-pixel text rendering, the data layout must be ARGB.  This can be
  // a little confusing: kCGImageAlphaPremultipliedFirst makes it so that the
  // most significant byte has the alpha component.  kCGBitmapByteOrder32Host is
  // Little Endian on Intel, which is the target.  Since it's Little Endian, the
  // bytes are actually stored as BGRA.  However, the data we're passing in
  // belongs to Go's image.RGBA type, which is basically Big Endian.  When a
  // glyph is drawn, the final step is to swap R and B's values in each pixel.
  CGContextRef ctx = CGBitmapContextCreate(data, width, height, 8, width * 4,
                                           colorspace,
                                           kCGImageAlphaPremultipliedFirst |
                                           kCGBitmapByteOrder32Host);

  CGContextSelectFont(ctx, [[curFont fontName] cStringUsingEncoding:NSUTF8StringEncoding], [curFont pointSize], kCGEncodingFontSpecific);
  CGContextSetTextDrawingMode(ctx, kCGTextFill);
  CGContextSetShouldAntialias(ctx, YES);
  CGContextSetShouldSmoothFonts(ctx, YES);
  CGContextSetShouldSubpixelPositionFonts(ctx, YES);
  CGContextSetShouldSubpixelQuantizeFonts(ctx, YES);


  return ctx;
}

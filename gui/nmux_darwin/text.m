#import "text.h"

static TextPattern *firstPattern = NULL;

static void patternReleaseCallback(void *info) {}


// Glyph drawing yoinked from MacVim
// Reference: https://github.com/macvim-dev/macvim/blob/fd114bd2897540c4dd764ca359b061231684ce50/src/MacVim/MMCoreTextView.m
static inline size_t gatherGlyphs(CGGlyph glyphs[], size_t length) {
  size_t gc = 0;
  for (int i = 0, p = 0; i < length; i++) {
    if (glyphs[i] != 0) {
      gc++;
      glyphs[p++] = glyphs[i];
    }
  }
  return gc;
}

void drawTextInContext(CGContextRef ctx, CTFontRef font, const unichar *chars,
                       CGGlyph *glyphs, CGPoint *positions, size_t length) {
  if (CTFontGetGlyphsForCharacters(font, chars, glyphs, length)) {
    length = gatherGlyphs(glyphs, length);
    CTFontDrawGlyphs(font, glyphs, positions, length, ctx);
    return;
  }

  CGGlyph *g = glyphs;
  CGGlyph *ge = g + length;
  CGPoint *p = positions;
  const unichar *c = chars;

  while (glyphs < ge) {
    if (*g) {
      bool surrogatePair = false;
      while (*g && g < ge) {
        if (CFStringIsSurrogateHighCharacter(*c)) {
          surrogatePair = true;
          g += 2;
          c += 2;
        } else {
          g++;
          c++;
        }
        p++;
      }

      size_t count = g - glyphs;
      if (surrogatePair) {
        count = gatherGlyphs(glyphs, count);
      }
      CTFontDrawGlyphs(font, glyphs, positions, count, ctx);
    } else {
      while (0 == *g && g < ge) {
        if (CFStringIsSurrogateHighCharacter(*c)) {
          g += 2;
          c += 2;
        } else {
          g++;
          c++;
        }
        p++;
      }

      size_t count = c - chars;
      size_t try_count = count;
      CTFontRef fallback = nil;
      while (fallback == nil && try_count > 0) {
        fallback = nmux_GetFontForChars(chars, try_count, font);
        if (fallback == nil) {
          try_count /= 2;
        }
      }

      if (fallback == nil) {
        break;
      }

      drawTextInContext(ctx, fallback, chars, glyphs, positions, try_count);

      c -= count - try_count;
      g -= count - try_count;
      p -= count - try_count;
    }

    if (glyphs == g) {
      break;
    }

    chars = c;
    glyphs = g;
    positions = p;
  }
}


// This is used for pattern fills in drawRect:
static inline void drawTextPattern(void *info, CGContextRef ctx) {
  TextPattern *tp = (TextPattern *)info;
  CGSize cellSize = nmux_CellSize();
  CGRect bounds = CGRectMake(0, 0, cellSize.width,
                             cellSize.height);
  CGContextSetFillColorWithColor(ctx, CGRGB(tp->attrs.bg));
  CGContextFillRect(ctx, bounds);

  if (tp->c != ' ') {
    CGGlyph glyphs;
    CTFontRef font = nmux_CurrentFont();
    CGPoint positions = CGPointMake(nmux_InitialCharPos(font),
                                    nmux_FontDescent(font));
    unichar c = tp->c;

    CGContextSetFillColorWithColor(ctx, CGRGB(tp->attrs.fg));
    CGContextSetTextDrawingMode(ctx, kCGTextFill);
    drawTextInContext(ctx, font, &c, &glyphs, &positions, 1);
  }
}


CGPatternRef getTextPatternLayer(TextPattern tp) {
  TextPattern *p = firstPattern;
  TextPattern *last = NULL;
  TextAttr pt;

  while(p != NULL) {
    last = p;
    pt = p->attrs;
    if (p->c == tp.c && pt.attrs == tp.attrs.attrs && pt.bg == tp.attrs.bg
        && pt.fg == tp.attrs.fg && pt.sp == tp.attrs.sp) {
      return p->pattern;
    }
    p = p->next;
  }

  TextPattern *np = calloc(1, sizeof(TextPattern));
  memcpy(np, &tp, sizeof(TextPattern));
  np->pattern = NULL;
  np->next = NULL;

  if (last != NULL) {
    last->next = np;
  } else {
    firstPattern = np;
  }

  CGPatternCallbacks cb;
  cb.drawPattern = &drawTextPattern;
  cb.releaseInfo = &patternReleaseCallback;
  cb.version = 0;

  CGRect bounds = CGRectZero;
  bounds.size = nmux_CellSize();
  np->pattern = CGPatternCreate((void *)np, bounds, CGAffineTransformIdentity,
                                bounds.size.width, bounds.size.height,
                                kCGPatternTilingConstantSpacing, YES, &cb);
  return np->pattern;
}

void textPatternClear() {
  TextPattern *p = firstPattern;
  TextPattern *n = NULL;

  while(p != NULL) {
    n = p->next;
    p->next = NULL;
    CGPatternRelease(p->pattern);
    free(p);
    p = n;
  }

  firstPattern = NULL;
}


/* vim: set ft=objc ts=2 sw=2 tw=80 et :*/

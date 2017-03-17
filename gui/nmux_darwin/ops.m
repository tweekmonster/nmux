#import "ops.h"


#pragma mark - DrawOp
@implementation DrawOp : NSObject

- (void)setDirtyX:(int)x y:(int)y w:(int)w h:(int)h {
  _bounds = NSMakeRect(x, y, w, h);
  NSSize cellSize = [nmux cellSize];
  _dirtyRect.origin.x = (CGFloat)x * cellSize.width;
  _dirtyRect.origin.y = (CGFloat)y * cellSize.height;
  _dirtyRect.size.width = (CGFloat)w * cellSize.width;
  _dirtyRect.size.height = (CGFloat)h * cellSize.height;
}
@end


#pragma mark - DrawTextOp
@implementation DrawTextOp
+ (DrawTextOp *)opWithText:(const char *)text x:(int)x y:(int)y
                     attrs:(TextAttr)attrs {
  DrawTextOp *op = [[DrawTextOp alloc] init];
  NSString *str = [NSString stringWithUTF8String:text];
  [op setText:str];
  [op setAttrs:attrs];
  [op setDirtyX:x y:y w:(int)[str length] h:1];
  return [op autorelease];
}

- (void)dealloc {
  if (_text != nil) {
    [_text release];
  }
  [super dealloc];
}

@end


#pragma mark - DrawRepeatedTextOp

static TextPattern *firstPattern = NULL;

@implementation DrawRepeatedTextOp
+ (DrawRepeatedTextOp *)opWithCharacter:(unichar)c length:(int)length x:(int)x
                                      y:(int)y attrs:(TextAttr)attrs {
  DrawRepeatedTextOp *op = [[DrawRepeatedTextOp alloc] init];
  [op setCharacter:c];
  [op setLength:length];
  [op setAttrs:attrs];
  [op setDirtyX:x y:y w:length h:1];
  return [op autorelease];
}


// This is used for pattern fills in drawRect:
static inline void drawTextPattern(void *info, CGContextRef ctx) {
  TextPattern *tp = (TextPattern *)info;
  CGRect bounds = CGRectMake(0, 0, [nmux cellSize].width,
                             [nmux cellSize].height);
  CGContextSetFillColorWithColor(ctx, CGRGB(tp->attrs.bg));
  CGContextFillRect(ctx, bounds);

  if (tp->c != ' ') {
    CGGlyph glyphs;
    CGPoint positions = CGPointMake([nmux firstCharPos], [nmux descent]);
    unichar c = tp->c;
    CTFontGetGlyphsForCharacters((CTFontRef)[nmux font], &c, &glyphs, 1);

    CGContextSetFillColorWithColor(ctx, CGRGB(tp->attrs.fg));
    CGContextSetTextDrawingMode(ctx, kCGTextFill);
    CTFontDrawGlyphs((CTFontRef)[nmux font], &glyphs, &positions, 1, ctx);
  }
}


static void patternReleaseCallback(void *info) {}


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
  bounds.size = [nmux cellSize];
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

@end
#pragma mark - ClearOp
@implementation ClearOp
+ (ClearOp *)opWithBg:(int32_t)bg {
  ClearOp *op = [[ClearOp alloc] init];
  [op setAttrs:(TextAttr){0, 0, bg, 0}];
  // No need to mark the dirty region since this clears the whole window.
  // TODO: Clear ops should send a rectangle for multiple nvim processes.
  return [op autorelease];
}

@end


#pragma mark - ScrollOp
@implementation ScrollOp
+ (ScrollOp *)opWithBg:(int32_t)bg delta:(int)delta top:(int)top
                bottom:(int)bottom left:(int)left right:(int)right {
  ScrollOp *op = [[ScrollOp alloc] init];
  [op setAttrs:(TextAttr){0, 0, bg, 0}];
  [op setDelta:delta];
  [op setDirtyX:left y:top w:(right-left)+1 h:(bottom-top)+1];
  return [op autorelease];
}

@end


#pragma mark - CursorOp
@implementation CursorOp
+ (CursorOp *)opWithX:(int)x y:(int)y character:(UniChar)character
                attrs:(TextAttr)attrs {
  CursorOp *op = [[CursorOp alloc] init];
  [op setCharacter:character];
  [op setAttrs:attrs];
  [op setDirtyX:x y:y w:1 h:1];
  return [op autorelease];
}

@end

/* vim: set ft=objc ts=2 sw=2 et :*/

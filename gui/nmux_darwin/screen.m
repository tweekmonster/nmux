#import "screen.h"
#import "bridge.h"
#import "nmux.h"
#import "misc.h"
#import "cgo_extern.h"

static inline NSMutableString * mouse_name(NSEvent *event) {
  NSString *name;
  switch ([event buttonNumber]) {
    case 0:
      name = @"Left";
      break;
    case 1:
      name = @"Right";
      break;
    case 2:
      name = @"Middle";
      break;
    default:
      return nil;
  }

  return [NSMutableString stringWithString:name];
}


@implementation NmuxScreen

- (instancetype)initWithFrame:(NSRect)frameRect {
  if (self = [super initWithFrame:frameRect]) {
    drawLock = [NSLock new];
    [self setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];
    flushOps = [NSMutableArray new];
  }
  return self;
}

- (Mode)state {
  return _state;
}

- (void)setState:(Mode)state {
  cursorUpdate = (state & ModeRedraw) != ModeRedraw;
  _state = state;
}

- (void)dealloc {
  if (runChars != NULL) {
    free(runChars);
  }

  if (runGlyphs != NULL) {
    free(runGlyphs);
  }

  if (runPositions != NULL) {
    free(runPositions);
  }

  [flushOps removeAllObjects];
  [flushOps release];

  [super dealloc];
}

- (BOOL)isFlipped {
  return NO;
}

#pragma mark - NSResponder
- (void)keyDown:(NSEvent *)event {
  // We only care about the keyDown event.
  BOOL shift = ([event modifierFlags] & NSShiftKeyMask) == NSShiftKeyMask;
  BOOL ctrl = ([event modifierFlags] & NSControlKeyMask) == NSControlKeyMask;
  BOOL alt = ([event modifierFlags] & NSAlternateKeyMask) == NSAlternateKeyMask;
  BOOL host = ([event modifierFlags] & NSCommandKeyMask) == NSCommandKeyMask;

  NSString *key = [event charactersIgnoringModifiers];
  unichar c = [key characterAtIndex:0];

  switch ([event keyCode]) {
    case kVK_F1:
      key = @"F1";
      break;
    case kVK_F2:
      key = @"F2";
      break;
    case kVK_F3:
      key = @"F3";
      break;
    case kVK_F4:
      key = @"F4";
      break;
    case kVK_F5:
      key = @"F5";
      break;
    case kVK_F6:
      key = @"F6";
      break;
    case kVK_F7:
      key = @"F7";
      break;
    case kVK_F8:
      key = @"F8";
      break;
    case kVK_F9:
      key = @"F9";
      break;
    case kVK_F10:
      key = @"F10";
      break;
    case kVK_F11:
      key = @"F11";
      break;
    case kVK_F12:
      key = @"F12";
      break;
    case kVK_F13:
      key = @"F13";
      break;
    case kVK_F14:
      key = @"F14";
      break;
    case kVK_F15:
      key = @"F15";
      break;
    case kVK_F16:
      key = @"F16";
      break;
    case kVK_F17:
      key = @"F17";
      break;
    case kVK_F18:
      key = @"F18";
      break;
    case kVK_F19:
      key = @"F19";
      break;
    case kVK_F20:
      key = @"F20";
      break;

    case kVK_Escape:
      key = @"Esc";
      break;
    case kVK_ForwardDelete:
      key = @"Del";
      break;
    case kVK_Return:
      key = @"CR";
      break;
    case kVK_Tab:
      key = @"Tab";
      break;

    case kVK_Help:
      // XXX: Huh?
      key = @"Ins";
      break;
    case kVK_Delete:
      key = @"BS";
      break;
    case kVK_Home:
      key = @"Home";
      break;
    case kVK_End:
      key = @"End";
      break;
    case kVK_PageUp:
      key = @"PageUp";
      break;
    case kVK_PageDown:
      key = @"PageDown";
      break;

    case kVK_UpArrow:
      key = @"Up";
      break;
    case kVK_DownArrow:
      key = @"Down";
      break;
    case kVK_LeftArrow:
      key = @"Left";
      break;
    case kVK_RightArrow:
      key = @"Right";
      break;

    case kVK_ANSI_KeypadDivide:
      key = @"kDivide";
      break;
    case kVK_ANSI_KeypadMultiply:
      key = @"kMultiply";
      break;
    case kVK_ANSI_KeypadMinus:
      key = @"kMinus";
      break;
    case kVK_ANSI_KeypadPlus:
      key = @"kPlus";
      break;
    case kVK_ANSI_KeypadEnter:
      key = @"kEnter";
      break;
    case kVK_ANSI_KeypadDecimal:
      key = @"kPoint";
      break;
    case kVK_ANSI_Keypad1:
      key = @"k1";
      break;
    case kVK_ANSI_Keypad2:
      key = @"k2";
      break;
    case kVK_ANSI_Keypad3:
      key = @"k3";
      break;
    case kVK_ANSI_Keypad4:
      key = @"k4";
      break;
    case kVK_ANSI_Keypad5:
      key = @"k5";
      break;
    case kVK_ANSI_Keypad6:
      key = @"k6";
      break;
    case kVK_ANSI_Keypad7:
      key = @"k7";
      break;
    case kVK_ANSI_Keypad8:
      key = @"k8";
      break;
    case kVK_ANSI_Keypad9:
      key = @"k9";
      break;
    case kVK_ANSI_Keypad0:
      key = @"k0";
      break;

    default:
      if (c >= ' ' && c <= '~') {
        switch (c) {
          case '<':
          case ',':
          case '>':
          case '.':
            shift = NO;
            ctrl = NO;
            alt = NO;
            host = NO;
            break;

          case ' ':
            key = @"Space";
            break;

          case '|':
            key = @"Bar";
            shift = NO;
            break;

          case '\\':
            key = @"Bslash";
            break;

          case '!':
          case '@':
          case '#':
          case '$':
          case '%':
          case '^':
          case '&':
          case '*':
          case '(':
          case ')':
          case '`':
          case '~':
          case '-':
          case '_':
          case '=':
          case '+':
          case '[':
          case '{':
          case '}':
          case ']':
          case ':':
          case ';':
          case '\'':
          case '"':
          case '/':
          case '?':
            shift = NO;
            break;

          default:
            if (c >= 'A' && c <= 'Z') {
              shift = NO;
            }
            break;
        }
      }
      break;
  }

  if (host) {
    key = [@"D-" stringByAppendingString:key];
  }

  if (alt) {
    key = [@"A-" stringByAppendingString:key];
  }

  if (ctrl) {
    key = [@"C-" stringByAppendingString:key];
  }

  if (shift) {
    key = [@"S-" stringByAppendingString:key];
  }

  if ([key length] > 1) {
    key = [NSString stringWithFormat:@"<%@>", key];
  }

  inputEvent((uintptr_t)self, (char *)[key UTF8String]);
}

- (BOOL)performKeyEquivalent:(NSEvent *)event {
  // Allows certain key combinations to be interpreted
  // (e.g. <c-Esc>, <a-Right>, etc).
  if (([event modifierFlags] & NSCommandKeyMask) == NSCommandKeyMask) {
    UniChar c = [[event characters] characterAtIndex:0];
    switch (c) {
      case 'q':
      case 'w':
      case 'n':
      case '`':
        return NO;
    }
  }

  [self keyDown:event];
  return YES;
}

- (NSPoint)mouseCoords {
  NSPoint coords = [[self window] mouseLocationOutsideOfEventStream];
  NSSize cellSize = [nmux cellSize];
  coords.x = floor(coords.x / cellSize.width);
  coords.y = floor((NSHeight([self bounds]) - coords.y) / cellSize.height);
  return coords;
}

- (void)dispatchMouseEvent:(NSString *)mouseKey {
  NSPoint coords = [self mouseCoords];
  if ([mouseKey hasSuffix:@"Drag"] && NSEqualPoints(coords, lastMouseCoords)) {
    return;
  }

  lastMouseCoords = coords;
  mouseKey = [NSString stringWithFormat:@"<%@><%d,%d>",
              mouseKey, (int)coords.x, (int)coords.y];
  inputEvent((uintptr_t)self, (char *)[mouseKey UTF8String]);
}

- (void)mouseDown:(NSEvent *)event {
  NSMutableString *name = mouse_name(event);
  if (name == nil) {
    return;
  }

  [name appendString:@"Mouse"];
  [self dispatchMouseEvent:name];
}

- (void)mouseUp:(NSEvent *)event {
  NSMutableString *name = mouse_name(event);
  if (name == nil) {
    return;
  }

  [name appendString:@"Release"];
  [self dispatchMouseEvent:name];
}

- (void)mouseDragged:(NSEvent *)event {
  NSMutableString *name = mouse_name(event);
  if (name == nil) {
    return;
  }

  [name appendString:@"Drag"];
  [self dispatchMouseEvent:name];
}

- (void)rightMouseDown:(NSEvent *)event {
  [self mouseDown:event];
}

- (void)rightMouseUp:(NSEvent *)event {
  [self mouseUp:event];
}

- (void)rightMouseDragged:(NSEvent *)event {
  [self mouseDragged:event];
}

- (void)otherMouseDown:(NSEvent *)event {
  [self mouseDown:event];
}

- (void)otherMouseUp:(NSEvent *)event {
  [self mouseUp:event];
}

- (void)otherMouseDragged:(NSEvent *)event {
  [self mouseDragged:event];
}

- (void)scrollWheel:(NSEvent *)event {
  if ([event deltaY] < 0) {
    [self dispatchMouseEvent:@"ScrollWheelDown"];
  } else {
    [self dispatchMouseEvent:@"ScrollWheelUp"];
  }
}

#pragma mark - Nmux Drawing

- (void)setGridSize:(NSSize)size {
  _grid = size;
  NSSize frameSize = NSSizeMultiply(size, [nmux cellSize]);
  NSRect winFrame = [[self window] frameRectForContentRect:NSMakeRect(0, 0, frameSize.width, frameSize.height)];
  winFrame.origin = [[self window] frame].origin;
  [[self window] setDelegate:nil];
  [[self window] setFrame:winFrame display:YES];
  [[self window] setDelegate:self];

  transform = CGAffineTransformMakeScale(1, -1);
  transform = CGAffineTransformTranslate(transform, 0, -NSHeight([self bounds]));

  cursorTransform = CGAffineTransformMakeScale(1, -1);
  cursorTransform = CGAffineTransformTranslate(cursorTransform, 0,
                                               -[nmux cellSize].height);

  [drawLock lock];
  [self lockFocus];
  CGContextRef ctx = [[NSGraphicsContext currentContext] graphicsPort];
  CGLayerRef newLayer = CGLayerCreateWithContext(ctx, frameSize, NULL);
  if (screenLayer != NULL) {
    CGContextDrawLayerAtPoint(ctx, CGPointApplyAffineTransform(CGPointZero, transform), screenLayer);
    CGLayerRelease(screenLayer);
  }
  screenLayer = newLayer;

  if (cursorLayer == NULL) {
    cursorLayer = CGLayerCreateWithContext(ctx, [nmux cellSize], NULL);
  } else {
    CGSize cursorSize = CGLayerGetSize(cursorLayer);
    if (!CGSizeEqualToSize(cursorSize, [nmux cellSize])) {
      CGLayerRelease(cursorLayer);
      cursorLayer = CGLayerCreateWithContext(ctx, [nmux cellSize], NULL);
    }
  }

  [self unlockFocus];
  [drawLock unlock];
}

- (void)drawTextContext:(CGContextRef)ctx op:(DrawTextOp *)op rect:(CGRect)rect {
  CGContextSaveGState(ctx);
  CGContextClipToRect(ctx, rect);
  CGContextSetFillColorWithColor(ctx, CGRGB([op attrs].bg));
  CGContextFillRect(ctx, rect);

  size_t runLength = (size_t)[[op text] length];

  if (runLength > runMaxLength) {
    if (runMaxLength > 0) {
      free(runChars);
      free(runGlyphs);
      free(runPositions);
    }

    runMaxLength = runLength;
    runChars = calloc(runMaxLength, sizeof(unichar));
    runGlyphs = calloc(runMaxLength, sizeof(CGGlyph));
    runPositions = calloc(runMaxLength, sizeof(CGPoint));
  }

  [[op text] getCharacters:runChars range:NSMakeRange(0, runLength)];
  NSSize cellSize = [nmux cellSize];
  CGFloat descent = [nmux descent];

  CTFontGetGlyphsForCharacters((CTFontRef)[nmux font], runChars,
                               runGlyphs, runLength);

  for (size_t i = 0; i < runLength; i++) {
    runPositions[i] = CGPointMake(i * cellSize.width, 0);
  }

  CGContextSetFillColorWithColor(ctx, CGRGB([op attrs].fg));
  CGContextSetTextDrawingMode(ctx, kCGTextFill);

  CGPoint o = rect.origin;
  o.x += [nmux firstCharPos];
  o.y += descent;
  CGContextTranslateCTM(ctx, o.x, o.y);
  CTFontDrawGlyphs((CTFontRef)[nmux font], runGlyphs, runPositions, runLength, ctx);
  CGContextTranslateCTM(ctx, -o.x, -o.y);
  CGContextRestoreGState(ctx);
}

- (void)drawTextPatternInContext:(CGContextRef)ctx op:(DrawRepeatedTextOp *)op rect:(CGRect)rect {
  TextPattern tp;
  tp.c = [op character];
  tp.attrs = [op attrs];
  tp.pattern = NULL;
  tp.next = NULL;
  CGPatternRef pattern = getTextPatternLayer(tp);
  CGFloat a = 1.0;

  CGContextSaveGState(ctx);
  CGColorSpaceRef ps = CGColorSpaceCreatePattern(NULL);
  CGContextSetFillColorSpace(ctx, ps);
  CGContextSetFillPattern(ctx, pattern, &a);
  CGContextFillRect(ctx, rect);
  CGColorSpaceRelease(ps);
  CGContextRestoreGState(ctx);
}

- (void)scrollInContext:(CGContextRef)ctx op:(ScrollOp *)op rect:(CGRect)rect {
  CGFloat offset = [op delta] * [nmux cellSize].height;
  CGRect clip = rect;
  clip.origin.y += offset;
  clip.size.height -= ABS(offset);

  CGContextSaveGState(ctx);
  CGContextSetShouldAntialias(ctx, NO);

  CGContextClipToRect(ctx, rect);
  CGContextTranslateCTM(ctx, 0, offset);
  CGContextDrawLayerAtPoint(ctx, CGPointZero, screenLayer);
  CGContextTranslateCTM(ctx, 0, -offset);

  if (offset < 0) {
    clip.origin.y = rect.origin.y + clip.size.height;
  } else {
    clip.origin.y = rect.origin.y;
  }

  clip.size.height = fabs(offset);
  CGContextSetFillColorWithColor(ctx, CGRGB([op attrs].bg));
  CGContextFillRect(ctx, clip);

  CGContextRestoreGState(ctx);
}

- (void)drawRect:(NSRect)dirtyRect {
  [drawLock lock];
  CGContextRef ctx = [[NSGraphicsContext currentContext] graphicsPort];

  if (screenLayer != NULL) {
    CGSize screenSize = CGLayerGetSize(screenLayer);
    CGContextDrawLayerAtPoint(ctx, CGPointMake(0, NSHeight([self bounds]) - screenSize.height), screenLayer);
  }

  if (cursorLayer != NULL) {
    CGContextDrawLayerAtPoint(ctx, cursorRect.origin, cursorLayer);
  }

  [super drawRect:dirtyRect];
  didFlush = NO;
  [drawLock unlock];
}

- (void)addDrawOp:(id)op {
  [drawLock lock];
  didFlush = NO;
  [flushOps addObject:op];
  [drawLock unlock];
}

- (BOOL)needsDisplay {
  [drawLock lock];
  BOOL display = didFlush;
  [drawLock unlock];
  return display;
}

- (void)flushDrawOps:(unichar)character pos:(NSPoint)cursorPos attrs:(TextAttr)attrs {
  [drawLock lock];

  CGContextRef ctx = CGLayerGetContext(screenLayer);
  CGContextSaveGState(ctx);
  CGContextSetShouldSubpixelPositionFonts(ctx, YES);
  CGContextSetShouldSubpixelQuantizeFonts(ctx, YES);

  for (DrawOp *op in flushOps) {
    if ([op isKindOfClass:[ClearOp class]]) {
      [op setDirtyX:0 y:0 w:_grid.width h:_grid.height];
    }

    CGRect rect = CGRectApplyAffineTransform([op dirtyRect], transform);

    if ([op isKindOfClass:[ClearOp class]]) {
      CGContextSetFillColorWithColor(ctx, CGRGB([op attrs].bg));
      CGContextFillRect(ctx, rect);
      textPatternClear();
    } else if ([op isKindOfClass:[DrawTextOp class]]) {
      [self drawTextContext:ctx op:(DrawTextOp *)op rect:rect];
    } else if ([op isKindOfClass:[DrawRepeatedTextOp class]]) {
      [self drawTextPatternInContext:ctx op:(DrawRepeatedTextOp *)op rect:rect];
    } else if ([op isKindOfClass:[ScrollOp class]]) {
      [self scrollInContext:ctx op:(ScrollOp *)op rect:rect];
    }

    [self setNeedsDisplayInRect:rect];
  }

  CGContextRestoreGState(ctx);
  [flushOps removeAllObjects];

  if (!NSEqualRects(cursorRect, NSZeroRect)) {
    [self setNeedsDisplayInRect:cursorRect];
  }

  if (cursorUpdate) {
    ctx = CGLayerGetContext(cursorLayer);
    CGContextSaveGState(ctx);
    CGContextSetShouldSubpixelPositionFonts(ctx, YES);
    CGContextSetShouldSubpixelQuantizeFonts(ctx, YES);

    NSString *cursorChar = [NSString stringWithCharacters:&character length:1];
    DrawTextOp *op = [DrawTextOp opWithText:[cursorChar UTF8String] x:0 y:0 attrs:attrs];
    CGSize cellSize = [nmux cellSize];
    CGPoint pos = CGPointMake(cursorPos.x * cellSize.width,
                              (cursorPos.y ) * cellSize.height);
    CGRect rect = CGRectApplyAffineTransform([op dirtyRect], cursorTransform);
    [self drawTextContext:ctx op:op rect:rect];
    cursorRect.origin = pos;
    cursorRect.size = [nmux cellSize];
    cursorRect = CGRectApplyAffineTransform(cursorRect, transform);
    [self setNeedsDisplayInRect:cursorRect];

    CGContextRestoreGState(ctx);
  }

  didFlush = YES;

  [drawLock unlock];
}


#pragma mark - Window Delegate
- (NSSize)windowWillResize:(NSWindow *)sender toSize:(NSSize)frameSize {
  NSRect contentFrame = [sender contentRectForFrameRect:NSMakeRect(0, 0, frameSize.width, frameSize.height)];
  NSSize minSize = NSSizeMultiply([nmux minGridSize], [nmux cellSize]);
  NSSize newSize = [nmux fitGrid:contentFrame.size];

  if (newSize.width < minSize.width) {
    newSize.width = minSize.width;
  }

  if (newSize.height < minSize.height) {
    newSize.height = minSize.height;
  }

  newSize.height += frameSize.height - NSHeight(contentFrame);
  return newSize;
}

- (void)windowDidMove:(NSNotification *)notification {
  NSRect frame = [[self window] frame];
  [nmux setLastWinFrame:frame];
  winMoved((uintptr_t)self, (int)NSMinX(frame), (int)NSMinY(frame));
}

- (void)windowDidResize:(NSNotification *)notification {
  NSRect frame = [self bounds];
  [nmux setLastWinFrame:frame];
  NSSize newGrid = NSSizeDivide(frame.size, [nmux cellSize]);
  winResized((uintptr_t)self, (int)NSWidth(frame), (int)NSHeight(frame),
             (int)newGrid.width, (int)newGrid.height);
#ifndef NMUX_CGO
  [self setGridSize:NSSizeDivide([self bounds].size, [nmux cellSize])];
  spam(self);
#endif
}

- (void)windowDidBecomeKey:(NSNotification *)notification {
  [nmux setLastWinFrame:[[self window] frame]];
  winFocused((uintptr_t)self);
}

- (void)windowDidResignKey:(NSNotification *)notification {
  winFocusLost((uintptr_t)self);
}

- (void)windowWillClose:(NSNotification *)notification {
  winClosed((uintptr_t)self);
}

- (NSSize)window:(NSWindow *)window willUseFullScreenContentSize:(NSSize)proposedSize {
  return [self windowWillResize:window toSize:proposedSize];
}

@end

/* vim: set ft=objc ts=2 sw=2 et :*/

#import <Cocoa/Cocoa.h>
#import <AppKit/AppKit.h>
#import <Carbon/Carbon.h>

// Any modifications must be able to run without Go since it's much easier to
// prototype in Xcode.

#ifndef NMUX_CGO

#define appStarted() NSLog(@"appStarted")
#define appStopped() NSLog(@"appStopped")
#define appHidden() NSLog(@"appHidden")
#define inputEvent(win, key) NSLog(@"inputEvent: %lu - %s", win, key)
#define winMoved(win, x, y, w, h) NSLog(@"winMoved: %lu - Pos: %ldx%ld, Size: %ldx%ld", win, x, y, w, h)
#define winClosed(win) NSLog(@"winClosed: %lu", win)
#define winFocused(win) NSLog(@"winFocused: %lu", win)
#define winFocusLost(win) NSLog(@"winFocusLost: %lu", win)

// Don't use GCD when running standalone.
#define DISPATCH_S(block) (block)()
#define DISPATCH_A(block) (block)()

#else

#import "_cgo_export.h"
#define DISPATCH_S(block) dispatch_sync(dispatch_get_main_queue(), (block))
#define DISPATCH_A(block) dispatch_async(dispatch_get_main_queue(), (block))

#endif

#pragma mark - Misc {{{1
#define RGB(c) [NSColor colorWithCalibratedRed:(CGFloat)(((c) >> 16) & 0xff) / 255 \
                                         green:(CGFloat)(((c) >> 8) & 0xff) / 255 \
                                          blue:(CGFloat)((c) & 0xff) / 255 \
                                         alpha:1]

typedef struct {
  uint8_t attrs;
  int32_t fg;
  int32_t bg;
  int32_t sp;
} TextAttr;

static NSFont *font;
static NSSize cellSize;
static NSSize minWinSize;
static NSRect lastWinFrame;


#pragma mark - Interfaces {{{1
@interface nmux : NSObject // {{{2
+ (void)setup;

+ (void)setFontFamily:(NSString *)name size:(CGFloat)size;
+ (NSFont *)font;
+ (CGFloat)descent;
+ (CGFloat)firstCharPos;
+ (NSSize)cellSize;
+ (NSSize)minWinSize;

+ (void)setLastWinFrame:(NSRect)frame;
+ (NSRect)lastWinFrame;
@end

@interface DrawOp : NSObject

@property (atomic, readonly) NSRect dirtyRect;
@property (atomic) TextAttr attrs;

- (void)setDirtyX:(int)x y:(int)y w:(int)w h:(int)h;
@end


@interface DrawTextOp : DrawOp; // {{{2

@property (atomic, retain) NSString *text;

+ (DrawTextOp *)opWithText:(const char *)text x:(int)x y:(int)y
                     attrs:(TextAttr)attrs;
@end


@interface ClearOp : DrawOp; // {{{2

+ (ClearOp *)opWithBg:(int32_t)bg;
@end


@interface ScrollOp : DrawOp; // {{{2

@property (atomic) int delta;
@property (atomic) NSRect region;

+ (ScrollOp *)opWithBg:(int32_t)bg delta:(int)delta top:(int)top
                bottom:(int)bottom left:(int)left right:(int)right;
@end

@interface CursorOp : DrawOp; // {{{2

@property (atomic) UniChar character;

+ (CursorOp *)opWithX:(int)x y:(int)y character:(UniChar)character
                attrs:(TextAttr)attrs;
@end

@interface NmuxScreen : NSView <NSWindowDelegate> // {{{2
{
  NSMutableArray *drawOps;
  NSMutableArray *pendingDrawOps;
  CGLayerRef screenLayer;
  NSPoint lastMouseCoords;

  unichar *runChars;
  CGGlyph *runGlyphs;
  CGPoint *runPositions;
  size_t runMaxLength;
}

- (void)addDrawOp:(DrawOp *)op;
- (void)flushDrawOps;
@end


@interface AppDelegate : NSObject <NSApplicationDelegate> // {{{2
- (void)newWindow:(id)action;
@end


#pragma mark - Helper Functions {{{1
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

static inline void menu_item(NSMenu *menu, NSString *title, SEL action,
                                  NSString *key) {
  NSMenuItem *item = [[[NSMenuItem alloc] initWithTitle:title action:action
                                          keyEquivalent:key] autorelease];
  [menu addItem:item];
}

static inline NSMenu * create_app_menu() {// {{{
  NSString *appName = [[NSProcessInfo processInfo] processName];
  NSMenu *menubar = [[NSMenu new] autorelease];
  NSMenuItem *topItem = [[NSMenuItem new] autorelease];
  [menubar addItem:topItem];

  NSMenu *submenu = [[NSMenu new] autorelease];
  menu_item(submenu, @"New Window", @selector(newWindow:), @"n");
  menu_item(submenu, @"Close", @selector(performClose:), @"w");
  menu_item(submenu, [NSString stringWithFormat:@"Quit %@", appName],
            @selector(terminate:), @"q");
  [topItem setSubmenu:submenu];

  topItem = [[[NSMenuItem alloc] initWithTitle:@"Window"
                                        action:nil
                                 keyEquivalent:@""] autorelease];
  submenu = [[[NSMenu alloc] initWithTitle:@"Window"] autorelease];
  menu_item(submenu, @"Minimize", @selector(performMiniaturize:), @"m");
  menu_item(submenu, @"Zoom", @selector(performZoom:), @"");
  [submenu addItem:[NSMenuItem separatorItem]];
  menu_item(submenu, @"Bring All to Front", @selector(arrangeInFront:), @"");

  [topItem setSubmenu:submenu];
  [menubar addItem:topItem];

  return menubar;
}// }}}

#pragma mark - Functions {{{1
void startApp() {// {{{
  @autoreleasepool{
    [NSHelpManager setContextHelpModeActive:NO];
    [NSApplication sharedApplication];
    [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

    NSMenu *menubar = create_app_menu();
    [NSApp setMainMenu:menubar];

    [nmux setup];

    AppDelegate* delegate = [[[AppDelegate alloc] init] autorelease];
    [NSApp setDelegate:delegate];
    [NSApp run];
  }
}// }}}

void stopApp() {
  DISPATCH_A(^{
    [NSApp terminate:nil];
  });
}

uintptr_t newWindow(int width, int height) {// {{{
  __block NmuxScreen *view;

  DISPATCH_S(^{
    NSInteger style;

    style = NSTitledWindowMask;
    style |= NSResizableWindowMask;
    style |= NSMiniaturizableWindowMask;
    style |= NSClosableWindowMask;

    NSRect rect = [nmux lastWinFrame];

    if (width != 0) {
      rect.size.width = (CGFloat)width;
    }

    if (height != 0) {
      rect.size.height = (CGFloat)height;
    }

    NSWindow *window = [[NSWindow alloc]
                        initWithContentRect:rect
                                  styleMask:style
                                    backing:NSBackingStoreBuffered
                                      defer:NO];
    [window setTitle:@"nmux"];
    [window setDisplaysWhenScreenProfileChanges:YES];

    NSWindow *active = [NSApp keyWindow];
    NSPoint pos = NSZeroPoint;

    if (active != nil) {
      NSRect frame = [active frame];
      pos = frame.origin;
      pos.x += 20;
      pos.y += frame.size.height - 20;
    }

    [window cascadeTopLeftFromPoint:pos];
    [window setAcceptsMouseMovedEvents:YES];

    view = [[[NmuxScreen alloc] init] autorelease];
    [window setDelegate:view];
    [window setBackgroundColor:[NSColor blackColor]];
    [window setContentView:view];
    [window makeFirstResponder:view];
    [window makeKeyAndOrderFront:nil];

    NSLog(@"Window: %@", NSStringFromRect([[window contentView] bounds]));
  });

  return (uintptr_t)view;
}// }}}

void drawText(uintptr_t view, const char *text, int x, int y, uint8_t attrs,
              int32_t fg, int32_t bg, int32_t sp) {
  TextAttr t;
  t.attrs = attrs;
  t.fg = fg;
  t.bg = bg;
  t.sp = sp;
  [(NmuxScreen *)view addDrawOp:[DrawTextOp opWithText:text x:x y:y attrs:t]];
}

void flush(uintptr_t view) {
  [(NmuxScreen *)view flushDrawOps];
}

#ifndef NMUX_CGO
void spam(NmuxScreen *view) {
  uint32_t mx = (uint32_t)(NSWidth([view bounds]) / [nmux cellSize].width);
  uint32_t my = (uint32_t)(NSHeight([view bounds]) / [nmux cellSize].height);

#define rand_color() (int32_t)arc4random_uniform(0xffffff)

  for (int i = 0; i < 1000; i++) {
    int x = arc4random_uniform(mx);
    int y = arc4random_uniform(my);
    TextAttr t = {0, rand_color(), rand_color(), rand_color()};
    drawText((uintptr_t)view, "Hello, World!", x, y, t);
  }
  [view flushDrawOps];
}
#endif

#pragma mark - nmux {{{1
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
                 weight:NSFontWeightRegular
                   size:size];
  }

  cellSize.width = ceil([f maximumAdvancement].width);
  cellSize.height = ceil([f ascender] + ABS([f descender]));
  minWinSize.width = cellSize.width * 80;
  minWinSize.height = cellSize.height * 20;

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

+ (NSSize)minWinSize {
  return minWinSize;
}

+ (void)setLastWinFrame:(NSRect)frame {
  lastWinFrame = frame;
}

+ (NSRect)lastWinFrame {
  if (NSEqualRects(lastWinFrame, NSZeroRect)) {
    lastWinFrame.size = minWinSize;
    NSRect screenRect = [[NSScreen mainScreen] visibleFrame];
    lastWinFrame.origin.x = 0;
    lastWinFrame.origin.y = NSHeight(screenRect) - NSHeight(lastWinFrame);
  }
  return lastWinFrame;
}

- (void)dealloc {
  if (font != nil) {
    [font release];
  }

  [super dealloc];
}

@end


#pragma mark - DrawOp {{{1
@implementation DrawOp : NSObject
- (void)setDirtyX:(int)x y:(int)y w:(int)w h:(int)h {
  NSSize cellSize = [nmux cellSize];
  _dirtyRect.origin.x = (CGFloat)x * cellSize.width;
  _dirtyRect.origin.y = (CGFloat)y * cellSize.height;
  _dirtyRect.size.width = (CGFloat)w * cellSize.width;
  _dirtyRect.size.height = (CGFloat)h * cellSize.height;
}
@end


#pragma mark - DrawTextOp {{{1
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


#pragma mark - ClearOp {{{1
@implementation ClearOp
+ (ClearOp *)opWithBg:(int32_t)bg {
  ClearOp *op = [[ClearOp alloc] init];
  [op setAttrs:(TextAttr){0, 0, bg, 0}];
  // No need to mark the dirty region since this clears the whole window.
  // TODO: Clear ops should send a rectangle for multiple nvim processes.
  return [op autorelease];
}

@end


#pragma mark - ScrollOp {{{1
@implementation ScrollOp
+ (ScrollOp *)opWithBg:(int32_t)bg delta:(int)delta top:(int)top
                bottom:(int)bottom left:(int)left right:(int)right {
  ScrollOp *op = [[ScrollOp alloc] init];
  [op setAttrs:(TextAttr){0, 0, bg, 0}];
  [op setDelta:delta];
  [op setDirtyX:left y:top w:right-left h:bottom-top];
  return [op autorelease];
}

@end


#pragma mark - CursorOp {{{1
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


#pragma mark - NmuxSurface {{{1
@implementation NmuxScreen

- (instancetype)initWithFrame:(NSRect)frameRect {
  if (self = [super initWithFrame:frameRect]) {
    [self setAutoresizingMask:NSViewWidthSizable|NSViewHeightSizable];
    drawOps = [[NSMutableArray alloc] init];
    pendingDrawOps = [[NSMutableArray alloc] init];
  }
  return self;
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

  [drawOps release];
  [pendingDrawOps release];

  [super dealloc];
}

- (BOOL)isFlipped {
  return YES;
}

#pragma mark - NSResponder {{{2
- (void)keyDown:(NSEvent *)event {// {{{
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
}// }}}

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
    [self dispatchMouseEvent:@"ScrollWheelUp"];
  } else {
    [self dispatchMouseEvent:@"ScrollWheelDown"];
  }
}

#pragma mark - Nmux Drawing {{{2
- (void)addDrawOp:(id)op {
  @synchronized(pendingDrawOps) {
    [pendingDrawOps addObject:op];
  }
}

- (void)flushDrawOps {
  @synchronized(pendingDrawOps) {
    @synchronized(drawOps) {
      NSRect frame = [self frame];

      for (DrawOp *op in pendingDrawOps) {
        [drawOps addObject:op];
        if ([op isKindOfClass:[ClearOp class]]) {
          [op setDirtyX:NSMinX(frame) y:NSMinY(frame)
                      w:NSWidth(frame) h:NSHeight(frame)];
          [self setNeedsDisplay:YES];
        } else {
          [self setNeedsDisplayInRect:[op dirtyRect]];
        }
      }
      [pendingDrawOps removeAllObjects];
    }
  }
}

- (void)drawRect:(NSRect)dirtyRect {
  @synchronized(drawOps) {
    CGContextRef ctx = [[NSGraphicsContext currentContext] CGContext];
    if (screenLayer == NULL) {
      screenLayer = CGLayerCreateWithContext(ctx, [self bounds].size, NULL);
    } else {
      CGSize screenSize = CGLayerGetSize(screenLayer);
      CGSize winSize = (CGSize)[self bounds].size;
      if (screenSize.width < winSize.width || screenSize.height < winSize.height) {
        CGLayerRef newLayer = CGLayerCreateWithContext(ctx, winSize, NULL);
        CGContextRef layerCtx = CGLayerGetContext(newLayer);
        CGContextDrawLayerAtPoint(layerCtx, CGPointZero, screenLayer);
        CGLayerRelease(screenLayer);
        screenLayer = newLayer;
      }
    }

    if ([drawOps count] != 0) {
      CGContextRef layerCtx = CGLayerGetContext(screenLayer);
      // A this point, drawOps should contain everything that matches the rects
      // obtained from getRectsBeingDrawn:count.
      CGContextSaveGState(layerCtx);

      for (DrawOp *op in drawOps) {
        CGFloat tx = NSMinX([op dirtyRect]);
        CGFloat ty = NSMinY([op dirtyRect]);

        CGContextTranslateCTM(layerCtx, tx, ty);
        NSColor *bg = RGB([op attrs].bg);

        CGContextSetFillColorWithColor(layerCtx, [bg CGColor]);
        CGContextFillRect(layerCtx, CGRectMake(0, 0, NSWidth([op dirtyRect]), NSHeight([op dirtyRect])));

        if ([op isKindOfClass:[ClearOp class]]) {

        } else if ([op isKindOfClass:[ScrollOp class]]) {

        } else if ([op isKindOfClass:[DrawTextOp class]]) {
          DrawTextOp *o = (DrawTextOp *)op;
          NSUInteger runLength = [[o text] length];

          NSColor *fg = RGB([op attrs].fg);

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

          [[o text] getCharacters:runChars];
          NSSize cellSize = [nmux cellSize];
          CGFloat descent = [nmux descent];

          CTFontGetGlyphsForCharacters((CTFontRef)[nmux font], runChars,
                                       runGlyphs, runLength);

          for (int i = 0; i < runLength; i++) {
            runPositions[i] = CGPointMake(i * cellSize.width, descent);
          }

          CGContextSetFillColorWithColor(layerCtx, [fg CGColor]);
          CGContextSetTextDrawingMode(layerCtx, kCGTextFill);

          CGContextSetTextMatrix(layerCtx, CGAffineTransformMakeScale(1.0, -1.0));
          CGFloat p = ceilf([nmux firstCharPos]);
          CGContextTranslateCTM(layerCtx, p, cellSize.height);

          CTFontDrawGlyphs((CTFontRef)[nmux font], runGlyphs, runPositions, runLength, layerCtx);
          CGContextTranslateCTM(layerCtx, -p, -cellSize.height);
          CGContextSetTextMatrix(layerCtx, CGAffineTransformIdentity);
        }

        CGContextTranslateCTM(layerCtx, -tx, -ty);
      }

      CGContextRestoreGState(layerCtx);
      [drawOps removeAllObjects];
    }

    CGContextDrawLayerAtPoint(ctx, CGPointZero, screenLayer);
  }

  [super drawRect:dirtyRect];
}

#pragma mark - Window Delegate {{{2
- (NSSize)windowWillResize:(NSWindow *)sender toSize:(NSSize)frameSize {
  NSSize minWinSize = [nmux minWinSize];
  NSSize cellSize = [nmux cellSize];
  NSRect contentFrame = [sender contentRectForFrameRect:NSMakeRect(0, 0, frameSize.width, frameSize.height)];
  frameSize.width = MAX(minWinSize.width,
                        NSWidth(contentFrame) - fmod(NSWidth(contentFrame),
                                               cellSize.width));
  CGFloat titleBar = frameSize.height - NSHeight(contentFrame);
  frameSize.height = titleBar + MAX(minWinSize.height,
                         NSHeight(contentFrame) - fmod(NSHeight(contentFrame),
                                                 cellSize.height));
  return frameSize;
}

- (void)windowDidMove:(NSNotification *)notification {
  NSRect frame = [[self window] frame];
  [nmux setLastWinFrame:frame];
  winMoved((uintptr_t)self, (NSInteger)NSMinX(frame), (NSInteger)NSMinY(frame),
              (NSInteger)NSWidth(frame), (NSInteger)NSHeight(frame));
}

- (void)windowDidResize:(NSNotification *)notification {
  [self windowDidMove:notification];
#ifndef NMUX_CGO
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

#pragma mark - AppDelegate {{{1
@implementation AppDelegate // {{{2

- (void)newWindow:(id)action {
  newWindow(0, 0);
}

- (void)applicationWillUpdate:(NSNotification *)notification {
  // ???
}

- (void)applicationDidFinishLaunching:(NSNotification *)notification {
  NSLog(@"Launching");
  [[NSRunningApplication currentApplication]
   activateWithOptions:NSApplicationActivateAllWindows|NSApplicationActivateIgnoringOtherApps];

#ifndef NMUX_CGO
  NmuxScreen *view = (NmuxScreen *)newWindow(0, 0);
  spam(view);
#endif

  appStarted();
}

- (void)applicationWillTerminate:(NSNotification *)notification {
  NSLog(@"Terminate");
  appStopped();
}

- (void)applicationWillHide:(NSNotification *)notification {
  NSLog(@"Hiding");
  appHidden();
}

@end

#ifndef NMUX_CGO
int main(int argc, char *argv[]) {
  startApp();
  return 0;
}
#endif
// }}}

/* vim: set ft=objc ts=2 sw=2 tw=80 fdm=marker cms=//\ %s et :*/

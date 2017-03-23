#import "bridge.h"

static inline void menu_item(NSMenu *menu, NSString *title, SEL action,
                                  NSString *key) {
  NSMenuItem *item = [[[NSMenuItem alloc] initWithTitle:title action:action
                                          keyEquivalent:key] autorelease];
  [menu addItem:item];
}

static inline NSMenu * create_app_menu() {
  NSString *appName = [[NSProcessInfo processInfo] processName];
  NSMenu *menubar = [[NSMenu new] autorelease];
  NSMenuItem *topItem = [[NSMenuItem new] autorelease];
  [menubar addItem:topItem];

  NSMenu *submenu = [[NSMenu new] autorelease];
  menu_item(submenu, @"New Window", @selector(applicationMenuSelected:), @"n");
  menu_item(submenu, @"Close", @selector(applicationMenuSelected:), @"w");
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
}


void startApp() {
  @autoreleasepool{
    [NSHelpManager setContextHelpModeActive:NO];
    [NSApplication sharedApplication];
    [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

    NSMenu *menubar = create_app_menu();
    [NSApp setMainMenu:menubar];

    nmux_Init();

    AppDelegate* delegate = [[[AppDelegate alloc] init] autorelease];
    [NSApp setDelegate:delegate];
    [NSApp run];
  }
}

void stopApp() {
  DISPATCH_A(^{
    [NSApp terminate:nil];
  });
}

uintptr_t newWindow(int width, int height) {
  __block NmuxScreen *view;

  DISPATCH_S(^{
    NSInteger style;

    style = NSTitledWindowMask;
    style |= NSResizableWindowMask;
    style |= NSMiniaturizableWindowMask;
    style |= NSClosableWindowMask;

    NSSize winGrid = NSMakeSize(width, height);
    NSSize minGrid = nmux_MinGridSize();
    NSRect rect = nmux_LastWindowFrame();

    if (width == 0) {
      winGrid.width = (int)minGrid.width;
    }

    if (height == 0) {
      winGrid.height = (int)minGrid.height;
    }

    rect.size = NSSizeMultiply(winGrid, nmux_CellSize());

    NSWindow *window = [[NSWindow alloc]
                        initWithContentRect:rect
                                  styleMask:style
                                    backing:NSBackingStoreNonretained
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
    [view setGridSize:winGrid];
    [window makeFirstResponder:view];
    [window makeKeyAndOrderFront:nil];

    NSLog(@"Window: %@", NSStringFromRect([[window contentView] bounds]));
  });

  return (uintptr_t)view;
}

void setGridSize(uintptr_t view, int cols, int rows) {
  DISPATCH_A(^{
    NmuxScreen *screen = (NmuxScreen *)view;
    [screen setGridSize:NSMakeSize(cols, rows)];
  });
}

void drawText(uintptr_t view, const char *text, int length, int index,
              uint8_t attrs, int32_t fg, int32_t bg, int32_t sp) {
  NSString *str = [NSString stringWithUTF8String:text];

  DISPATCH_A((^{
    NmuxScreen *screen = (NmuxScreen *)view;
    int x = index % (int)[screen grid].width;
    int y = index / (int)[screen grid].width;
    TextAttr t;

    t.attrs = attrs;
    t.fg = fg;
    t.bg = bg;
    t.sp = sp;

    [screen addDrawOp:[DrawTextOp opWithText:str x:x y:y attrs:t]];
  }));
}

void drawRepeatedText(uintptr_t view, unichar character, int length, int index,
                      uint8_t attrs, int32_t fg, int32_t bg, int32_t sp) {
  DISPATCH_A(^{
    NmuxScreen *screen = (NmuxScreen *)view;
    int x = index % (int)[screen grid].width;
    int y = index / (int)[screen grid].width;
    TextAttr t;

    t.attrs = attrs;
    t.fg = fg;
    t.bg = bg;
    t.sp = sp;
    [screen addDrawOp:[DrawRepeatedTextOp opWithCharacter:character length:length x:x y:y attrs:t]];
  });
}

void scrollScreen(uintptr_t view, int delta, int top, int bottom, int left,
                  int right, int32_t bg) {
  DISPATCH_A(^{
    NmuxScreen *screen = (NmuxScreen *)view;
    [screen addDrawOp:[ScrollOp opWithBg:bg delta:delta top:top
                                  bottom:bottom left:left right:right]];
  });
}

void clearScreen(uintptr_t view, int32_t bg) {
  DISPATCH_A(^{
    NmuxScreen *screen = (NmuxScreen *)view;
    [screen addDrawOp:[ClearOp opWithBg:bg]];
  });
}

void flush(uintptr_t view, int mode, int x, int y, const char *character,
           int width, uint8_t attrs, int32_t fg, int32_t bg, int32_t sp) {
  NSString *str = [NSString stringWithUTF8String:character];

  DISPATCH_A(^{
    NmuxScreen *screen = (NmuxScreen *)view;
    [screen setState:(Mode)mode];

    TextAttr ta;

    ta.attrs = attrs;
    ta.fg = fg;
    ta.bg = bg;
    ta.sp = sp;

    NSPoint pos;
    pos.x = x;
    pos.y = y;
    [screen flushDrawOps:str charWidth:width pos:pos attrs:ta];
  });
}

void setTitle(uintptr_t view, const char *title) {
  NSString *str = [NSString stringWithUTF8String:title];
  DISPATCH_A(^{
    NmuxScreen *screen = (NmuxScreen *)view;
    [[screen window] setTitle:[@"nmux: "
      stringByAppendingString:str]];
  });
}

void setIcon(uintptr_t view, const char *icon) {

}

void bell(uintptr_t view, bool visual) {
  DISPATCH_A(^{
   [(NmuxScreen *)view beep:visual];
  });
}

void getCellSize(int *x, int *y) {
  NSSize cellSize = nmux_CellSize();
  *x = (int)cellSize.width;
  *y = (int)cellSize.height;
}

#ifndef NMUX_CGO
void spam(NmuxScreen *view) {
  int mx = (int)([view grid].width);
  int my = (int)([view grid].height);

#define rand_color() (int32_t)arc4random_uniform(0xffffff)

  NSString *message = @"Emoji 👍 test 💩 Test String™ beep boop 🤖";
  const char *messageChars = [message UTF8String];
  int message_width = (int)[message length];
  int message_len = (int)[message lengthOfBytesUsingEncoding:NSUTF8StringEncoding];

  int y = 0;
  int p = 0x112233, b, ci = 0;
  int i1, i2;
  unichar c;

  // Ugly but consistent color cycling
#define nextp() (p + (p << ((y % 3) * 7))) % 0xffffff

  while (y < my) {
    p = nextp();
    b = 0xffffff - p;

    c = (unichar)'A' + (ci % 26);
    ci++;
    i1 = y * mx;
    drawText((uintptr_t)view, messageChars, message_len, i1, 0,
             p, b, rand_color());
    i2 = i1 + ((mx - message_width) / 2);
    i1 += message_width;

    p = nextp();
    b = 0xffffff - p;
    drawRepeatedText((uintptr_t)view, c, i2 - i1 - 2, i1 + 1, 0, p, b, p);

    p = nextp();
    b = 0xffffff - p;
    drawText((uintptr_t)view, messageChars, message_len, i2, 0, p, b, p);

    c = (unichar)'A' + (ci % 26);
    ci++;
    i1 = i2 + message_width;
    i2 = (y * mx + mx) - message_width;
    p = nextp();
    b = 0xffffff - p;
    drawRepeatedText((uintptr_t)view, c, i2 - i1 - 2, i1 + 1, 0, p, b, p);

    p = nextp();
    b = 0xffffff - p;
    drawText((uintptr_t)view, messageChars, message_len, i2, 0, p, b, p);
    y++;
  }

  flush((uintptr_t)view, 0, 1, 1, "X", 1, 0, 0xffffff, 0, 0xffffff);
}

int main(int argc, char *argv[]) {
  startApp();
  return 0;
}
#endif

/* vim: set ft=objc ts=2 sw=2 et :*/

#ifndef MISC_H
#define MISC_H
#import <Cocoa/Cocoa.h>

typedef struct {
  uint8_t attrs;
  int32_t fg;
  int32_t bg;
  int32_t sp;
} TextAttr;


typedef struct TP {
  unichar c;
  TextAttr attrs;
  CGPatternRef pattern;
  struct TP *next;
} TextPattern;


typedef enum _Mode {
  ModeBusy    = 1 << 0,
  ModeMouseOn = 1 << 1,
  ModeNormal  = 1 << 2,
  ModeInsert  = 1 << 3,
  ModeReplace = 1 << 4,
  ModeRedraw  = 1 << 5,
  ModeEnd     = 1 << 6,
} Mode;


static inline NSSize NSSizeMultiply(NSSize s1, NSSize s2) {
  s1.width *= s2.width;
  s1.height *= s2.height;
  return s1;
}

static inline NSSize NSSizeDivide(NSSize s1, NSSize s2) {
  s1.width /= s2.width;
  s1.height /= s2.height;
  return s1;
}


#define NSRGB(c) [NSColor colorWithDeviceRed:(CGFloat)(((c) >> 16) & 0xff) / 255 \
                                       green:(CGFloat)(((c) >> 8) & 0xff) / 255 \
                                        blue:(CGFloat)((c) & 0xff) / 255 \
                                       alpha:1]
#define CGRGB(c) [NSRGB(c) CGColor]
#endif /* ifndef  MISC_H */

/* vim: set ft=objc ts=2 sw=2 et :*/

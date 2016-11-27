#ifndef TEXT_DARWIN_H
#define TEXT_DARWIN_H
#include <Foundation/Foundation.h>
#include <AppKit/AppKit.h>

bool set_font(const char *, CGFloat, CGFloat *, CGFloat *, CGFloat *, CGFloat *);
bool get_glyph(UniChar, CGGlyph *);
CGContextRef create_bitmap_context(uint8_t *, size_t, size_t);
void draw_glyph(CGContextRef, CGGlyph, int, int, CGColorRef, CGColorRef);
#endif /* ifndef TEXT_DARWIN_H */

#ifndef TEXT_H
#define TEXT_H
#import <Cocoa/Cocoa.h>
#import "nmux.h"
#import "misc.h"

void drawTextInContext(CGContextRef, CTFontRef, const unichar *, CGGlyph *,
                       CGPoint *, size_t);
CGPatternRef getTextPatternLayer(TextPattern);
void textPatternClear(void);
#endif /* ifndef TEXT_H */

/* vim: set ft=objc ts=2 sw=2 tw=80 et :*/

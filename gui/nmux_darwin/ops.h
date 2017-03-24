#ifndef OPS_H
#define OPS_H
#import <Cocoa/Cocoa.h>
#import "nmux.h"

@interface DrawOp : NSObject
{
  NSRect _bounds;
  NSRect _dirtyRect;
}

@property (atomic, readonly) NSRect dirtyRect;
@property (atomic) TextAttr attrs;

- (void)setDirtyX:(int)x y:(int)y w:(int)w h:(int)h;
@end


@interface DrawTextOp : DrawOp
{
  NSString *_text;
  BOOL _cursor;
}
@property (atomic, retain) NSString *text;
@property (atomic) BOOL cursor;

+ (DrawTextOp *)opWithText:(NSString *)text x:(int)x y:(int)y
                     attrs:(TextAttr)attrs;
@end

@interface DrawRepeatedTextOp : DrawOp

@property (atomic) unichar character;
@property (atomic) int length;

+ (DrawRepeatedTextOp *)opWithCharacter:(unichar)c length:(int)length x:(int)x y:(int)y
                                  attrs:(TextAttr)attrs;

@end


@interface ClearOp : DrawOp

+ (ClearOp *)opWithBg:(int32_t)bg;
@end


@interface ScrollOp : DrawOp;

@property (atomic) int delta;
@property (atomic) NSRect region;

+ (ScrollOp *)opWithBg:(int32_t)bg delta:(int)delta top:(int)top
                bottom:(int)bottom left:(int)left right:(int)right;
@end

@interface CursorOp : DrawOp;

@property (atomic) UniChar character;

+ (CursorOp *)opWithX:(int)x y:(int)y character:(UniChar)character
                attrs:(TextAttr)attrs;
@end
#endif /* ifndef OPS_H */

/* vim: set ft=objc ts=2 sw=2 et :*/

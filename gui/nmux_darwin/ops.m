#import "ops.h"


#pragma mark - DrawOp
@implementation DrawOp : NSObject

- (void)setDirtyX:(int)x y:(int)y w:(int)w h:(int)h {
  _bounds = NSMakeRect(x, y, w, h);
  NSSize cellSize = nmux_CellSize();
  _dirtyRect.origin.x = (CGFloat)x * cellSize.width;
  _dirtyRect.origin.y = (CGFloat)y * cellSize.height;
  _dirtyRect.size.width = (CGFloat)w * cellSize.width;
  _dirtyRect.size.height = (CGFloat)h * cellSize.height;
}
@end


#pragma mark - DrawTextOp
@implementation DrawTextOp
+ (DrawTextOp *)opWithText:(NSString *)text x:(int)x y:(int)y
                     attrs:(TextAttr)attrs {
  DrawTextOp *op = [[DrawTextOp alloc] init];
  // Note: The width is not based on character sizes since nmux takes care of
  // character width spacing.
  [op setText:text];
  [op setAttrs:attrs];
  [op setDirtyX:x y:y w:(int)[text length] h:1];
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

// vim: set ft=objc ts=2 sw=2 et :

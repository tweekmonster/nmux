#ifndef NMUX_H
#define NMUX_H
#import <Cocoa/Cocoa.h>
#import "misc.h"

void nmux_Init(void);
CGSize nmux_CellSize(void);
CGSize nmux_MinGridSize(void);
CGSize nmux_FitToGrid(CGSize);
NSRect nmux_LastWindowFrame(void);
void nmux_SetLastWindowFrame(NSRect);
void nmux_SetFont(NSString *, CGFloat);
CTFontRef nmux_GetFontForChars(const unichar *, UniCharCount, CTFontRef);
CTFontRef nmux_CurrentFont(void);
CGFloat nmux_FontDescent(CTFontRef);
CGFloat nmux_InitialCharPos(CTFontRef);
#endif /* ifndef NMUX_H */

/* vim: set ft=objc ts=2 sw=2 et :*/

// +build darwin

package gui

// This allows us to split objc sources into smaller files.  Since app_darwin.go
// exports C functions, we can't include C files there.  This is explained in:
// https://golang.org/cmd/cgo/#hdr-C_references_to_Go
//
// For posterity:
//  > Using //export in a file places a restriction on the preamble: since it is
//  > copied into two different C output files, it must not contain any
//  > definitions, only declarations. If a file contains both definitions and
//  > declarations, then the two output files will produce duplicate symbols and
//  > the linker will fail. To avoid this, definitions must be placed in
//  > preambles in other files, or in C source files.

/*
#cgo CFLAGS: -x objective-c -D NMUX_CGO -mmacosx-version-min=10.8 -D__MAC_OS_X_VERSION_MAX_ALLOWED=1080
#cgo LDFLAGS: -framework Cocoa -framework AppKit

#import "nmux_darwin/bridge.m"
#import "nmux_darwin/ops.m"
#import "nmux_darwin/text.m"
#import "nmux_darwin/nmux.m"
#import "nmux_darwin/delegate.m"
#import "nmux_darwin/screen.m"
*/
import "C"

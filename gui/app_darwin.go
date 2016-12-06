package gui

/*
#cgo CFLAGS: -x objective-c -mmacosx-version-min=10.8 -D__MAC_OS_X_VERSION_MAX_ALLOWED=1080 -D NMUX_CGO
#cgo LDFLAGS: -framework Cocoa -framework AppKit

#include <stdint.h>

void startApp();
void stopApp();
uintptr_t newWindow(int, int);
void setGridSize(uintptr_t, int, int);
void drawText(uintptr_t, const char *, int, int, uint8_t, int32_t, int32_t, int32_t);
void flush(uintptr_t);
*/
import "C"
import (
	"runtime"
	"sync"

	"github.com/tweekmonster/nmux/screen"
	"github.com/tweekmonster/nmux/util"
)

func init() {
	runtime.LockOSThread()
}

var appCallback func()
var windowInputs map[uintptr]chan interface{}

type window struct {
	id     uintptr
	mu     sync.Mutex
	events chan interface{}
}

func (w *window) GetID() uintptr {
	return w.id
}

func (w *window) PutChar(c rune, attrs screen.CellAttrs) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return nil
}

func (w *window) PutString(s string, attrs screen.CellAttrs) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return nil
}

func (w *window) Scroll() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return nil
}

func (w *window) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return nil
}

func (w *window) SendEvent(event interface{}) {
	w.events <- event
}

func (w *window) NextEvent() interface{} {
	return <-w.events
}

func platformStart(callback func()) error {
	appCallback = callback
	C.startApp()
	return nil
}

func platformStop() {
	C.stopApp()
}

func platformNewWindow(width, height int) Window {
	win := &window{
		id:     uintptr(C.newWindow(C.int(width), C.int(height))),
		events: make(chan interface{}),
	}
	return win
}

//export appStarted
func appStarted() {
	go appCallback()
}

//export appStopped
func appStopped() {
}

//export appHidden
func appHidden() {
}

//export inputEvent
func inputEvent(id uintptr, key *C.char) {
	util.Debug("Key:", C.GoString(key))
	sendEvent(id, InputEvent(C.GoString(key)))
	C.drawText(C.uintptr_t(id), key, 1, 1, 0, 0xffffff, 0x00ff00, 0x000000)
	C.flush(C.uintptr_t(id))
}

//export winMoved
func winMoved(id uintptr, x, y, w, h int) {
	sendEvent(id, ResizeEvent{
		X:      x,
		Y:      y,
		Width:  w,
		Height: h,
	})
}

//export winClosed
func winClosed(id uintptr) {
	sendEvent(id, StateEvent("closed"))
}

//export winFocused
func winFocused(id uintptr) {
	sendEvent(id, StateEvent("focused"))
}

//export winFocusLost
func winFocusLost(id uintptr) {
	sendEvent(id, StateEvent("lostFocus"))
}

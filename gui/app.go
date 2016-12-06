package gui

import (
	"errors"

	"github.com/tweekmonster/nmux/screen"
)

var ErrNoApp = errors.New("no app available")
var ErrUnableToCreateWin = errors.New("unable to create window")

type Window interface {
	GetID() uintptr
	PutChar(c rune, attrs screen.CellAttrs) error
	PutString(s string, attrs screen.CellAttrs) error
	Scroll() error
	Close() error
	SendEvent(interface{})
	NextEvent() interface{}
}

type App struct {
	windows map[uintptr]Window
}

func (a *App) NewWindow(width, height int) (Window, error) {
	win := platformNewWindow(width, height)
	if win == nil {
		return nil, ErrUnableToCreateWin
	}

	a.windows[win.GetID()] = win
	return win, nil
}

func sendEvent(id uintptr, event interface{}) {
	if win, ok := app.windows[id]; ok {
		win.SendEvent(event)
	}
}

var app *App

func Start(loop func(*App)) error {
	app = &App{
		windows: make(map[uintptr]Window),
	}

	return platformStart(func() {
		loop(app)
		platformStop()
	})
}

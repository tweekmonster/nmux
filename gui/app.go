package gui

import (
	"errors"

	"github.com/tweekmonster/nmux/screen"
)

var ErrNoApp = errors.New("no app available")
var ErrUnableToCreateWin = errors.New("unable to create window")

type Window interface {
	GetID() uintptr
	PutString(s string, index int, attrs screen.CellAttrs) error
	PutRepeatedString(s rune, length, index int, attrs screen.CellAttrs) error
	SetGrid(cols, rows int) error
	Scroll(delta, top, bottom, left, right int, bg screen.Color) error
	Clear(bg screen.Color) error
	Flush(mode int, character string, width int, cursor screen.Vector2, attrs screen.CellAttrs) error
	SetTitle(title string) error
	SetIcon(icon string) error
	Bell(visual bool) error
	Close() error
	SendEvent(interface{})
	NextEvent() interface{}
	EventChannel() <-chan interface{}
}

type WindowEvent struct {
	Window
	Event interface{}
}

type ApplicationEvent struct {
	Event interface{}
}

type App struct {
	windows map[uintptr]Window
	events  chan interface{}
}

func (a *App) NewWindow(width, height int) (Window, error) {
	win := platformNewWindow(width, height)
	if win == nil {
		return nil, ErrUnableToCreateWin
	}

	a.windows[win.GetID()] = win
	return win, nil
}

func (a *App) GetWindow(id int) Window {
	// Stub.  The server will send an ID with the payload.
	for _, win := range a.windows {
		return win
	}
	return nil
}

func (a *App) CellSize() screen.Vector2 {
	return platformCellSize()
}

func (a *App) EventChannel() <-chan interface{} {
	return a.events
}

var app *App

func sendApplicationEvent(event interface{}) {
	app.events <- ApplicationEvent{Event: event}
}

func sendWindowEvent(id uintptr, event interface{}) {
	if win, ok := app.windows[id]; ok {
		app.events <- WindowEvent{Window: win, Event: event}
	}
}

func Start(loop func(*App)) error {
	app = &App{
		events:  make(chan interface{}),
		windows: make(map[uintptr]Window),
	}

	return platformStart(func() {
		go sendApplicationEvent(StateEvent("started"))
		loop(app)
		platformStop()
	})
}

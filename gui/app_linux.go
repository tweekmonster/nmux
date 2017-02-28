package gui

import "github.com/tweekmonster/nmux/screen"

func platformStart(callback func()) error {
	return ErrNoApp
}

func platformStop() {
}

func platformNewWindow(width, height int) Window {
	return nil
}

func platformCellSize() screen.Vector2 {
	return screen.Vector2{}
}

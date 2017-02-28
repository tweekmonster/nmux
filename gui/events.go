package gui

type StateEvent string

type InputEvent string

type ResizeEvent struct {
	Width      int
	Height     int
	GridWidth  int
	GridHeight int
}

type MoveEvent struct {
	X int
	Y int
}

type MenuEvent string

type NewWindowEvent struct {
	Window
}

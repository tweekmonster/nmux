package gui

func platformStart(callback func()) error {
	return ErrNoApp
}

func platformStop() {
}

func platformNewWindow(width, height int) Window {
	return nil
}

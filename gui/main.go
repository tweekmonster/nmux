package gui

import "github.com/tweekmonster/nmux/util"

func Main(addr string) {
	Start(func(a *App) {
		win, err := a.NewWindow(500, 400)
		if err != nil {
			panic(err)
		}

		for {
			e := win.NextEvent()
			util.Debug(e)
		}
	})
}

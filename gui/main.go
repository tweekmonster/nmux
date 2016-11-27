package gui

import (
	"image"
	"log"

	"github.com/tweekmonster/nmux/util"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

func Main() {
	ts := GetTypeSetter()
	if err := ts.SetFont("Menlo", 13); err != nil {
		log.Panic(err)
	}

	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(nil)
		if err != nil {
			log.Fatal(err)
		}

		defer w.Release()

		winSize := image.Point{256, 256}
		b, err := s.NewBuffer(winSize)
		if err != nil {
			log.Fatal(err)
		}
		defer b.Release()
		va := CreateVimAttr(0, 0x000000, 0xffffff, 0xffffff)

		var sz size.Event

		for {
			e := w.NextEvent()
			util.Debug("Event:", e)

			switch e := e.(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}

			case key.Event:
				if e.Code == key.CodeEscape {
					return
				}

			case paint.Event:
				w.Fill(sz.Bounds(), va.Bg, screen.Src)
				ts.PutString(b.RGBA(), 0, 0, "Hello, World!", va)
				w.Upload(image.Point{0, 0}, b, b.Bounds())
				w.Publish()

			case size.Event:
				sz = e

			case error:
				log.Print(e)
			}
		}
	})
}

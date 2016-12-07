package gui

import "github.com/tweekmonster/nmux/util"

func handleWindowEvent(client *Client, win Window, event interface{}) {
	switch e := event.(type) {
	case ResizeEvent:
		util.Print("RESIZE", e)
		client.Resize(e.GridWidth, e.GridHeight)
	case InputEvent:
		client.SendInput(string(e))
	}
}

func Main(addr string) {
	Start(func(a *App) {
		var client *Client

		for {
			select {
			case e, ok := <-a.EventChannel():
				if !ok {
					break
				}

				switch event := e.(type) {
				case ApplicationEvent:
					switch appEvent := event.Event.(type) {
					case StateEvent:
						if appEvent == "started" {
							var err error
							client, err = NewClient(addr, a)
							if err != nil {
								panic(err)
							}

							util.Print("Client:", client)
						}
					default:
						// util.Debug("AppEvent:", appEvent)
					}
				case WindowEvent:
					handleWindowEvent(client, event.Window, event.Event)

				default:
					util.Debug("WinEvent:", event)
				}
			}
		}
	})
}

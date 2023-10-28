package controllers

import (
	"fmt"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/gorilla/websocket"
	"time"
)

type WebSocketCtl struct {
	beego.Controller
}

var upgrade = websocket.Upgrader{}

func (w *WebSocketCtl) Get() {
	ws, err := upgrade.Upgrade(w.Ctx.ResponseWriter, w.Ctx.Request, nil)
	if err != nil {
		w.Ctx.ResponseWriter.Status = 400
		w.StopRun()
	}
	defer ws.Close()

	getMsgCh := make(chan []byte)
	go func(ch chan []byte) {
		for {
			t, p, err := ws.ReadMessage()
			if err != nil {
				fmt.Println("11111111", err)
				return
			}
			fmt.Println("get msg:", string(p), "msg type: ", t)
			ch <- p
		}
	}(getMsgCh)

	tk := time.NewTicker(1 * time.Second)

	n := 0
	for {
		select {
		case d := <-getMsgCh:
			if string(d) == "bye" {
				fmt.Println("client say bye.")
				return
			}
			err = ws.WriteJSON(fmt.Sprintf("i have got msg: %s", string(d)))
			if err != nil {
				fmt.Println("22222222")
				return
			}
		case <-tk.C:
			err = ws.WriteJSON(fmt.Sprintf("this is msg: %d", n))
			if err != nil {
				fmt.Println("33333333", err)
				return
			}
			n++
		}
	}
}

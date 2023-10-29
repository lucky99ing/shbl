package controllers

import (
	context2 "context"
	"fmt"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"shbl/coupler"
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
	defer close(getMsgCh)
	go func(ch chan []byte) {
		for {
			_, p, err := ws.ReadMessage()
			if err != nil {
				fmt.Println("11111111", err)
				return
			}
			//fmt.Println("get msg:", string(p), "msg type: ", t)
			ch <- p
		}
	}(getMsgCh)

	downCh := make(chan []byte, 1)
	upCh := make(chan []byte, 1)

	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("123456"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	//fd := int(os.Stdin.Fd())
	//termWidth, termHeight, err := terminal.GetSize(fd)
	//if err != nil {
	//	panic(err)
	//}

	t := coupler.Manager{
		Down1:      downCh,
		Up2:        upCh,
		IP:         "192.168.122.8",
		Port:       "22",
		SSHConfig:  sshConfig,
		TermWidth:  200,
		TermHeight: 100,
	}

	//tk := time.NewTicker(1 * time.Second)
	//defer tk.Stop()
	ctxSSH, cancel := context2.WithCancel(context2.Background())

	//n := 0
	go func() {
		for {
			select {
			case d := <-getMsgCh:
				if string(d) == "bye" {
					fmt.Println("client say bye.")
					cancel()
					return
				}
				downCh <- d
				//err = ws.WriteJSON(fmt.Sprintf("i have got msg: %s", string(d)))
				//if err != nil {
				//	fmt.Println("22222222")
				//	return
				//}
			case dat := <-upCh:
				ws.WriteJSON(string(dat))
				//case <-tk.C:
				//	err = ws.WriteJSON(fmt.Sprintf("this is msg: %d", n))
				//	if err != nil {
				//		fmt.Println("33333333", err)
				//		return
				//	}
				//	n++
			}
		}
	}()
	fmt.Println("cccccccccccc")
	err = t.StartSSH(ctxSSH)
	if err != nil {
		panic(err)
	}
	fmt.Println("----------end----------")
}

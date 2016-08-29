package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"golang.org/x/net/websocket"
	"sync/atomic"
)

var chMsg = make(chan msgType)
var aid uint64 = 0
var lConn = make(map[uint64]*websocket.Conn)
var cmdList = [...]string{
	`setPos`,
	`getWho`,
}

type msgType struct {
	id   uint64
	cmd  string
	data *simplejson.Json
}

func session() websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		id := atomic.AddUint64(&aid, 1)
		fmt.Println(`new connection #`, id)
		lConn[id] = ws

		msg := `{"cmd":"id","id":"` + fmt.Sprintf(`%d`, id) + `"}`
		websocket.Message.Send(ws, msg)

		for {
			msg = ``
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				fmt.Println(`#`, id, ` receive exit with`, err.Error())
				delete(lConn, id)
				return
			}

			// 忽略所有格式错误的消息
			json, err := simplejson.NewJson([]byte(msg))
			if err != nil {
				fmt.Println(`error json `, err)
				continue
			}

			cmd := json.Get(`cmd`).MustString()

			bIn := false
			for _, checkCmd := range cmdList {
				if checkCmd == cmd {
					bIn = true
					break
				}
			}
			if !bIn {
				fmt.Println(`unknown cmd `, cmd)
				continue
			}

			// 只返回格式正确的消息
			chMsg <- msgType{
				id:   id,
				cmd:  cmd,
				data: json,
			}
		}
	})
}

func loopRead() {
	for {
		msg := <-chMsg
		for id, Conn := range lConn {
			if id == msg.id {
				continue
			}
			d := msg.data
			js, _ := d.Encode()
			s := fmt.Sprintf(`uid: %d, text: %s`, msg.id, js)
			fmt.Println(s)

			d.Set(`id`, msg.id)
			js, _ = d.Encode()
			websocket.Message.Send(Conn, string(js[:]))
		}
	}
}

func main() {

	go loopRead()

	e := echo.New()
	e.GET("/ws", standard.WrapHandler(session()))
	e.Run(standard.New(":58888"))
}

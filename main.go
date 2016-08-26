package main

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	"golang.org/x/net/websocket"
)

var chMsg = make(chan string)
var handleSerial = int64(0)
var lConn = make(map[int64]*websocket.Conn)

func hello() websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		fmt.Println(ws)
		id := handleSerial + 1
		lConn[id] = ws
		for {
			msg := ""
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				fmt.Println(`receive exit with` + err.Error())
				delete(lConn, id)
				return
			}
			chMsg <- msg

			/*
				err = websocket.Message.Send(ws, `echo [`+msg+`]`)
				if err != nil {
					fmt.Println(`send exit with` + err.Error())
					return
				}
				fmt.Printf("%s\n", msg)
			*/
		}
	})
}

func loopRead() {
	for {
		msg := <-chMsg
		fmt.Println(`loopRead ` + msg)
	}
}

func main() {

	go loopRead()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/ws", standard.WrapHandler(hello()))
	e.Run(standard.New(":58888"))
}

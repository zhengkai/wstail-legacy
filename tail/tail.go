package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

var (
	configFileName = `tail_file.txt`
	lFileAllow     map[string]bool

	sessionMap    = make(map[uint64]*map[uint64]bool)
	sessionChan   = make(map[uint64]*chan uint64)
	tailBind      = make(chan sessionInfo, 1000)
	filePool      = make(map[uint64]*fileContent)
	fileMap       = make(map[string]*uint64)
	sessionSerial uint64
	fileSerial    uint64
	fileVer       uint64

	addr     = flag.String("addr", ":58888", "http service address")
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {

	go refreshConfig()
	go manager()

	http.HandleFunc(`/ws`, serveWs)

	fmt.Println(`WsTail started`)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

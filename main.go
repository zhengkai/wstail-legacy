package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var (
	whitelistFileName  string = `whitelist.txt`
	whitelistFileFinal string

	sessionMap    = make(map[uint64]*map[uint64]bool)
	sessionChan   = make(map[uint64]*chan uint64)
	tailBind      = make(chan sessionInfo, 1000)
	filePool      = make(map[uint64]*fileContent)
	fileMap       = make(map[string]uint64)
	httpListen    = `:58888`
	writeWait     time.Duration
	fileAllow     map[string]bool
	sessionSerial uint64
	fileSerial    uint64
	noopInterval  int64 = 25
	iWriteWait    int64 = 10

	upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {

	loadConfig()

	go refreshWhiteList()
	go manager()

	http.HandleFunc(`/ws/status`, statusPage)
	http.HandleFunc(`/ws/tail`, serveWs)
	http.HandleFunc(`/ws/config`, configPage)

	fmt.Println(`WsTail started`)
	err := http.ListenAndServe(httpListen, nil)
	if err != nil {
		panic("ListenAndServe fail: " + err.Error())
	}
}

func writeHttp(w http.ResponseWriter, s string) {
	w.Write([]byte(s))
}

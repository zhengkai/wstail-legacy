package main

import (
	"fmt"
	"gopkg.in/ini.v1"
	"log"
	"net/http"
	"time"
)

func loadConfig() {
	cfg, err := ini.Load(`tail.ini`)
	if err != nil {
		log.Fatal(err)
	}

	section := cfg.Section(``)

	// listen
	httpListen = section.Key(`listen`).MustString(httpListen)
	fmt.Println(`listen`, httpListen)

	// noopInterval
	noopInterval = section.Key(`noop-interval`).MustInt64(noopInterval)
	fmt.Println(`noopInterval`, noopInterval)

	// writeWait
	iWriteWait := section.Key(`write-wait`).MustInt64(iWriteWait)
	if iWriteWait < 1 {
		iWriteWait = 1
	}
	writeWait = time.Duration(iWriteWait) * time.Second
	fmt.Println(`writeWait`, writeWait)

	// whitelistFileName
	whitelistFileName = section.Key(`whitelist-file`).MustString(whitelistFileName)
	fmt.Println(`whitelist-file`, whitelistFileName)

	fmt.Println(``)
}

func configPage(w http.ResponseWriter, r *http.Request) {

	writeHttp(w, "Settings:\n\n")

	showVar(w, `listen`, httpListen)
	showVar(w, `noop-interval`, noopInterval)
	showVar(w, `write-wait`, iWriteWait)
	showVar(w, `whitelist-file`, whitelistFileName)

	writeHttp(w, "\n\nAllowed files:\n\n")

	for k, _ := range fileAllow {
		writeHttp(w, "\t"+k+"\n")
	}
}

func showVar(w http.ResponseWriter, k string, v interface{}) {
	writeHttp(w, fmt.Sprintf("\t%-14s = %v\n", k, v))
}

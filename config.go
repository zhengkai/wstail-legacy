package main

import (
	"fmt"
	"gopkg.in/ini.v1"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func loadConfig() {

	cfg := configInit()

	section := cfg.Section(``)

	// listen
	httpListen = section.Key(`listen`).MustString(httpListen)
	fmt.Println("\t", `listen`, httpListen)

	// noopInterval
	noopInterval = section.Key(`noop-interval`).MustInt64(noopInterval)
	fmt.Println("\t", `noopInterval`, noopInterval)

	// writeWait
	iWriteWait := section.Key(`write-wait`).MustInt64(iWriteWait)
	if iWriteWait < 1 {
		iWriteWait = 1
	}
	writeWait = time.Duration(iWriteWait) * time.Second
	fmt.Println("\t", `writeWait`, iWriteWait)

	// whitelistFileName
	whitelistFileName = section.Key(`whitelist-file`).MustString(whitelistFileName)
	fmt.Println("\t", `whitelist-file`, whitelistFileName)

	fmt.Println()
}

func configPage(w http.ResponseWriter, r *http.Request) {

	writeHttp(w, fmt.Sprintf("version: %s\n\n", version))

	writeHttp(w, "Settings:\n\n")

	showVar(w, `listen`, httpListen)
	showVar(w, `noop-interval`, noopInterval)
	showVar(w, `write-wait`, iWriteWait)
	showVar(w, `whitelist-file`, whitelistFileName)

	writeHttp(w, "\n\nWhitelist File:\n\n\t"+whitelistFileFinal+"\n")

	writeHttp(w, "\nAllowed files:\n\n")
	for k, _ := range fileAllow {
		writeHttp(w, "\t"+k+"\n")
	}

	writeHttp(w, "\nAllowed dirs:\n\n")
	for k, _ := range dirAllow {
		writeHttp(w, "\t"+k+"\n")
	}
}

func configInit() *ini.File {

	cfg := ini.Empty()

	baseName := `wstail.ini`
	bLoad := false
	file := ``

	dirList := make([]string, 2)

	dir, err := os.Getwd()
	if err == nil {
		dirList = append(dirList, dir)
	}
	dirList = append(dirList, os.Args[0])
	dirList = append(dirList, `/usr/etc/wstail.ini`)
	dirList = append(dirList, `/etc/wstail.ini`)

	for _, dir := range dirList {
		dir, err = filepath.Abs(filepath.Dir(dir))
		if err == nil {
			if dir != `/` {
				dir += `/`
			}
			file = dir + baseName
			cfgbak, err := ini.Load(file)
			if err == nil {
				cfg = cfgbak
				bLoad = true
				break
			}
		}
	}

	if !bLoad {
		file = `none`
	}

	fmt.Println(`config = `+file, "\n")

	return cfg
}

func showVar(w http.ResponseWriter, k string, v interface{}) {
	writeHttp(w, fmt.Sprintf("\t%-14s = %v\n", k, v))
}

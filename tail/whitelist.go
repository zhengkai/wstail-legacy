package main

import (
	"fmt"
	"github.com/hpcloud/tail"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func initWhiteList() (configFile string) {
	if whitelistFileName[0:1] == `/` {
		return whitelistFileName
	}

	appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	for _, path := range [...]string{appDir, pwd, `/etc`} {
		checkFile := path + `/` + whitelistFileName
		if _, err := os.Stat(checkFile); err == nil {
			configFile = checkFile
			break
		}
	}
	if configFile == `` {
		log.Fatal(`config file "` + whitelistFileName + `" not found`)
	}

	return
}

func refreshWhiteList() {

	configFile := initWhiteList()

	fmt.Println(`final whitelist file`, configFile)

	var iWait int64 = 3
	for {
		t, _ := tail.TailFile(configFile, tail.Config{Follow: true, Logger: tail.DiscardingLogger})
		fileAllow = make(map[string]bool)
		for line := range t.Lines {
			sLine := line.Text
			sLine = strings.TrimSpace(sLine)
			if sLine[0:1] != `/` {
				continue
			}
			fileAllow[sLine] = true
		}
		time.Sleep(time.Duration(iWait) * time.Second)
	}
}

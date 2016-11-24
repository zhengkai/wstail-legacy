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

func initWhiteList() (file string) {
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
			file = checkFile
			break
		}
	}
	if file == `` {
		log.Fatal(`config file "` + whitelistFileName + `" not found`)
	}

	return
}

func refreshWhiteList() {

	whitelistFileFinal = initWhiteList()

	fmt.Println(`final whitelist file`, whitelistFileFinal)

	fileAllow = make(map[string]bool)
	dirAllow = make(map[string]bool)

	var iWait int64 = 3
	for {
		t, _ := tail.TailFile(whitelistFileFinal, tail.Config{Follow: true, Logger: tail.DiscardingLogger})
		for line := range t.Lines {
			sLine := line.Text
			sLine = strings.TrimSpace(sLine)

			if !strings.HasPrefix(sLine, `/`) {
				continue
			}
			if strings.HasSuffix(sLine, `/`) {
				dirAllow[sLine] = true
			} else {
				fileAllow[sLine] = true
			}
		}
		time.Sleep(time.Duration(iWait) * time.Second)
	}
}

func checkFileInWhiteList(sFile string) bool {

	if strings.HasSuffix(sFile, `/`) {
		return false
	}

	var k string

	for k, _ = range fileAllow {
		if k == sFile {
			return true
		}
	}

	for k, _ = range dirAllow {
		if strings.HasPrefix(sFile, k) {
			return true
		}
	}

	return false
}

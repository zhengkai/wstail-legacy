package main

import (
	"github.com/hpcloud/tail"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func initConfig() (configFile string) {
	appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	for _, path := range [...]string{appDir, pwd, `/etc`} {
		checkConfigFile := path + `/` + configFileName
		if _, err := os.Stat(checkConfigFile); err == nil {
			configFile = checkConfigFile
			break
		}
	}
	if configFile == `` {
		log.Fatal(`config file "` + configFileName + `" not found`)
	}

	return
}

func refreshConfig() {

	configFile := initConfig()

	var iWait int64
	for {
		t, _ := tail.TailFile(configFile, tail.Config{Follow: true, Logger: tail.DiscardingLogger})
		iWait = 10
		lFileAllow = make(map[string]bool)
		for line := range t.Lines {
			iWait = 1
			sLine := line.Text
			sLine = strings.Trim(sLine, ` `)
			if sLine[0:1] != `/` {
				continue
			}
			lFileAllow[sLine] = true
		}
		time.Sleep(time.Duration(iWait) * time.Second)
	}
}

package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

func statusPage(w http.ResponseWriter, r *http.Request) {

	writeHttp(w, fmt.Sprintf("version: %s\n\n", version))

	t := time.Now().Round(time.Second).Sub(timeStart)
	writeHttp(w, fmt.Sprintf("uptime: %s\n", t))

	connNum := len(sessionChan)
	writeHttp(w, fmt.Sprintf("websocket connections: %d\n\n", connNum))

	writeHttp(w, fmt.Sprintf("data traffic out: %s\n", NumberToString(transOut, ',')))

	if connNum < 1 {
		return
	}

	writeHttp(w, "\nfile list:\n")
	for file, fid := range fileMap {
		writeHttp(w, fmt.Sprintf("%5d. %s\n", fid, file))

		sMap := sessionMap[fid]
		for sid, _ := range *sMap {
			writeHttp(w, fmt.Sprintf("\t\t%d\n", sid))
		}
	}

	writeHttp(w, "\nsession list:\n")
	for sid, _ := range sessionChan {
		writeHttp(w, fmt.Sprintf("\t%d\n", sid))
	}
}

func NumberToString(n uint64, sep rune) string {

	s := fmt.Sprintf(`%d`, n)

	startOffset := 0
	var buff bytes.Buffer

	l := len(s)

	commaIndex := 3 - ((l - startOffset) % 3)

	if commaIndex == 3 {
		commaIndex = 0
	}

	for i := startOffset; i < l; i++ {

		if commaIndex == 3 {
			buff.WriteRune(sep)
			commaIndex = 0
		}
		commaIndex++

		buff.WriteByte(s[i])
	}

	return buff.String()
}

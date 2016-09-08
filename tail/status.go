package main

import "fmt"
import "net/http"

func statusPage(w http.ResponseWriter, r *http.Request) {

	connNum := len(sessionChan)
	writeHttp(w, fmt.Sprintf("websocket connections: %d", connNum))
	if connNum < 1 {
		return
	}

	writeHttp(w, "\n\nfile list:\n")
	for file, fid := range fileMap {
		writeHttp(w, fmt.Sprintf("%5d. %s\n", fid, file))

		sMap := sessionMap[fid]
		for sid, _ := range *sMap {
			writeHttp(w, fmt.Sprintf("\t\t%d\n", sid))
		}
	}

	writeHttp(w, "\n\nsession list:\n")
	for sid, _ := range sessionChan {
		writeHttp(w, fmt.Sprintf("\t%d\n", sid))
	}
}

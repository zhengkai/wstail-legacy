package main

import (
	"bufio"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type sessionInfo struct {
	id   uint64
	file string
}

type sessionVer struct {
	ver  uint64
	line uint64
}

type fileContent struct {
	ver  uint64
	size int64
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// fmt.Println(`init upgrade fail`, err)
		return
	}
	err = ws.WriteMessage(websocket.TextMessage, []byte(`!test`))
	if err != nil {
		// fmt.Println(`init ws err`, err)
	}

	r.ParseForm()
	sFile := strings.TrimSpace(r.Form.Get(`file`))
	bValid := false
	for k, _ := range lFileAllow {
		if k == sFile {
			bValid = true
			break
		}
	}
	if !bValid {
		/*
			fmt.Println(`not allowed file`, sFile)
			for k, _ := range lFileAllow {
				fmt.Println(`>`, k)
			}
		*/
		ws.Close()
		return
	}

	// fmt.Println(`file`, sFile, `OK`)

	sid := atomic.AddUint64(&sessionSerial, 1)
	ch := make(chan uint64)
	sessionChan[sid] = &ch
	tailBind <- sessionInfo{
		id:   sid,
		file: sFile,
	}

	var ver uint64
	var offset int

	fmt.Println(`new session`, sid, sFile)

	go func() {
		for {
			fid := <-ch
			bLoop := send(sid, fid, sFile, &ver, &offset, ws)
			if !bLoop {
				delete(sessionChan, sid)
				delete(*sessionMap[fid], sid)
				break
			}
		}
	}()

	// go reporter(sid)
}

func send(sid uint64, fid uint64, file string, ver *uint64, offset *int, ws *websocket.Conn) bool {

	for {
		fc := filePool[fid]
		if fc == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		f, err := os.Open(file)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		defer f.Close()

		bReset := false
		if *ver != fc.ver {
			bReset = true
			*ver = fc.ver
			*offset = 0
		}

		buf := bufio.NewReader(f)
		if *offset > 0 {
			buf.Discard(*offset)
		}
		// fmt.Println(`offset = `, *offset)

		for {
			sbuf := make([]byte, 4096)
			n, err := buf.Read(sbuf)
			bEOF := false
			if err == io.EOF {
				if n < 1 {
					return true
				}
				*offset += n
				bEOF = true
			} else if err != nil {
				fmt.Println(`read err`, err)
				break
			} else {
				*offset += n
			}

			checkfc := *filePool[fid]
			if *ver != checkfc.ver {
				break
			}

			if bReset {
				bReset = false
				err = ws.WriteMessage(websocket.TextMessage, []byte(`!reset`))
			}

			sbuf = append([]byte{'>'}, sbuf...)
			err = ws.WriteMessage(websocket.TextMessage, sbuf)
			if err != nil {
				fmt.Println(`ws err`, err)
				return false
			}

			// fmt.Println(`ws send = "` + string(sbuf) + `"`)
			// ws.WriteMessage(websocket.BinaryMessage, []byte(`end`))
			if bEOF {
				break
			}
		}

		// fmt.Println(`sid =`, sid, `, fid =`, fid, `, ver =`, ver, `, offset =`, offset, `, fc =`, fc)
	}
	return true
}

func manager() {

	var fid uint64
	for {
		select {
		case sessInfo := <-tailBind:

			pfid := fileMap[sessInfo.file]
			if pfid == nil {
				fid = atomic.AddUint64(&fileSerial, 1)
				fileMap[sessInfo.file] = &fid
				go scan(fid, sessInfo.file)
			} else {
				fid = *pfid
			}

			list := sessionMap[fid]

			if list == nil {
				n := make(map[uint64]bool, 20)
				sessionMap[fid] = &n
			}

			n := *sessionMap[fid]
			n[sessInfo.id] = true

			_tend(sessInfo.id, fid)
		}
	}
}

func scan(fid uint64, file string) {

	ch := make(chan bool)

	go scanUpdate(fid, file, &ch)
	go watch(file, &ch)

	ch <- true
}

func scanUpdate(fid uint64, file string, ch *chan bool) {

	var ver uint64
	var size int64
	// var time int64

	for {
		t := <-*ch
		if !t {
			ver++
			size = 0
		}

		os.Stat(file)
		finfo, err := os.Stat(file)
		if err != nil {
			// fmt.Println(err)
			continue
		}

		if finfo.Size() == size {
			continue
		}
		if finfo.Size() < size {
			ver++
		}
		size = finfo.Size()
		// time = finfo.ModTime().UnixNano()

		filePool[fid] = &fileContent{
			ver:  ver,
			size: size,
		}

		go tend(fid)
	}
}

func tend(fid uint64) {
	// fmt.Println(`sessionMap`, sessionMap[fid])
	c := sessionMap[fid]
	if c == nil {
		return
	}

	for sid, _ := range *sessionMap[fid] {
		go _tend(sid, fid)
	}
}

func _tend(sid uint64, fid uint64) {
	ch := sessionChan[sid]
	if ch == nil {
		return
	}
	select {
	case *ch <- fid:
	default:
	}
}

func watch(file string, ch *chan bool) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	for {
		err := watcher.Add(file)
		if err != nil {
			continue
		}

		event := <-watcher.Events
		if event.Op&fsnotify.Write == fsnotify.Write {
			select {
			case *ch <- true:
			default:
			}
		}
	}
}

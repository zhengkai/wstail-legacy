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
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

type sessionInfo struct {
	id   uint64
	file string
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

	r.ParseForm()
	sFile := strings.TrimSpace(r.Form.Get(`file`))
	bValid := false
	for k, _ := range fileAllow {
		if k == sFile {
			bValid = true
			break
		}
	}
	if !bValid {
		err = wsWrite(ws, []byte(`!filedeny,file=`+sFile))
		ws.Close()
		return
	}
	// fmt.Println(`file`, sFile, `OK`)

	sid := atomic.AddUint64(&sessionSerial, 1)
	err = wsWrite(ws, []byte(fmt.Sprintf(`!connection,id=%d`, sid)))
	ch := make(chan uint64)
	// fmt.Println(`new chan`, &ch, ch)
	sessionChan[sid] = &ch
	tailBind <- sessionInfo{
		id:   sid,
		file: sFile,
	}

	ver, _ := strconv.ParseUint(r.Form.Get(`ver`), 10, 64)
	offset, _ := strconv.Atoi(r.Form.Get(`offset`))
	rCh := make(chan bool)

	fmt.Println(`new session`, sid, sFile, ver, offset)

	timeLastSent := time.Now().Unix()

	go func() {
		for {
			_, _, err := ws.ReadMessage()
			/*
				if strings.Contains(err.Error(), `close 1002`) {
					err = nil
				}
			*/
			if err == nil {
				select {
				case rCh <- true:
				default:
				}
			} else {
				rCh <- false
				// fmt.Println(`send false`, sid)
				// fmt.Println(a, b, err)
				return
			}
		}
	}()

	go func() {

		var rfid uint64
		bLoop := true

		// fmt.Println(`start loop`, ch)

		for {
			select {
			case fid := <-ch:
				timeLastSent = time.Now().Unix()
				if fid > 0 {
					rfid = fid
					bLoop = send(sid, fid, sFile, &ver, &offset, ws)
				} else {
					bLoop = sendNoop(ws)
				}
			case rc := <-rCh:
				timeLastSent = time.Now().Unix()
				// fmt.Println(`read stop loop`)
				bLoop = rc
			}
			// fmt.Println(`stop loop`, bLoop, sid)
			if !bLoop {
				break
			}
		}
		timeLastSent = 0
		delete(sessionChan, sid)
		if rfid > 0 {
			delete(*sessionMap[rfid], sid)
		}
		ws.Close()
	}()

	go sendNoopTimer(&timeLastSent, &ch)
}

func send(sid uint64, fid uint64, file string, ver *uint64, offset *int, ws *websocket.Conn) bool {

	bReset := false
	var tmpVer uint64 = 0

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

		if *ver != fc.ver {
			bReset = true
			tmpVer = fc.ver
			*offset = 0
		}

		buf := bufio.NewReader(f)
		if *offset > 0 {
			buf.Discard(*offset)
		}

		tmpCnt := 0

		prevBuf := []byte(``)

		for {
			sbuf := make([]byte, buffLen)
			n, err := buf.Read(sbuf)
			// fmt.Println(`read =`, n, err, tmpVer, sid)

			if err == io.EOF {
				if *offset > 0 && tmpVer > 0 {
					*ver = tmpVer
				}
				return true
			}

			if err != nil {
				break
			}

			*offset += n
			if len(prevBuf) > 0 {
				sbuf = append(prevBuf, sbuf[:n]...)
				prevBuf = []byte(``)
			} else if n < buffLen {
				sbuf = sbuf[:n]
			}

			if tmpVer > 0 {
				*ver = tmpVer
			}

			checkfc := *filePool[fid]
			if *ver != checkfc.ver {
				break
			}

			if bReset {
				bReset = false
				err = wsWrite(ws, []byte(fmt.Sprintf(`!reset,ver=%d`, *ver)))
			}

			if !utf8.Valid(sbuf) {
				pass := false
				slen := len(sbuf)
				for _, i := range [...]int{1, 2, 3, 4, 5, 6} {
					tbuf := sbuf[:slen-i]
					// fmt.Println(`tbuf len`, i, len(tbuf))
					if utf8.Valid(tbuf) {
						pass = true
						prevBuf = sbuf[slen-i:]
						sbuf = tbuf
						// fmt.Println(`utf8`, i, len(prevBuf))
						break
					}
				}
				if !pass {
					// fmt.Println(`sub fail`)
				}
			}

			sbuf = append([]byte{'>'}, sbuf...)

			tmpCnt += n
			// fmt.Println(`length =`, n, len(sbuf)-1)

			err = wsWrite(ws, sbuf)
			if err != nil {
				fmt.Println(`ws err`, err)
				return false
			}
		}
	}
	return true
}

func sendNoopTimer(t *int64, ch *chan uint64) {
	if *t < 1 {
		return
	}

	tNext := *t + noopInterval - time.Now().Unix()
	if tNext < 1 {
		select {
		case *ch <- 0:
		default:
		}
		tNext = noopInterval
	}
	time.Sleep(time.Second * time.Duration(tNext))
	go sendNoopTimer(t, ch)
}

func sendNoop(ws *websocket.Conn) bool {
	err := wsWrite(ws, []byte(`!noop`))
	if err != nil {
		return false
	}
	return true
}

func wsWrite(ws *websocket.Conn, msg []byte) error {
	transOut += uint64(len(msg))
	ws.SetWriteDeadline(time.Now().Add(writeWait))
	return ws.WriteMessage(websocket.TextMessage, msg)
}

func manager() {

	var fid uint64
	for {
		sessInfo := <-tailBind

		ftmp := fileMap[sessInfo.file]
		if ftmp > 0 {
			fid = ftmp
		} else {
			fileSerial++
			// fid = atomic.AddUint64(&fileSerial, 1)
			fid = fileSerial
			fileMap[sessInfo.file] = fid
			go scan(fid, sessInfo.file)
		}

		list := sessionMap[fid]
		if list == nil {
			n := make(map[uint64]bool, 20)
			n[sessInfo.id] = true
			sessionMap[fid] = &n
		} else {
			n := *list
			n[sessInfo.id] = true
		}

		go func() {
			*sessionChan[sessInfo.id] <- fid
			// fmt.Println(`sent chan`)
		}()
	}
}

func scan(fid uint64, file string) {

	ch := make(chan bool)

	go scanUpdate(fid, file, &ch)
	go watch(file, &ch)

	ch <- true
}

func scanUpdate(fid uint64, file string, ch *chan bool) {

	var ver uint64 = 1
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

		// if finfo.Size() == size {
		//	continue
		// }

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
	// fmt.Println(`_tend`, sid, fid, sessionChan[sid])
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

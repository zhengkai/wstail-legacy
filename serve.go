package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

type sessionInfo struct {
	id   uint64
	file string
}

type fileContent struct {
	ver  uint64
	size int64
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		// fmt.Println(`init upgrade fail`, err)
		return
	}

	r.ParseForm()
	sFile := strings.TrimSpace(r.Form.Get(`file`))

	bValid := false
	sFile, err = filepath.Abs(sFile)
	if err == nil {
		bValid = checkFileInWhiteList(sFile)
	}
	if !bValid {
		err = wsWrite(ws, []byte(`!filedeny,file=`+sFile))
		ws.Close()
		return
	}

	sid := atomic.AddUint64(&sessionSerial, 1)
	err = wsWrite(ws, []byte(fmt.Sprintf(`!connection,id=%d`, sid)))
	ch := make(chan uint64)
	sessionChan[sid] = &ch
	tailBind <- sessionInfo{
		id:   sid,
		file: sFile,
	}

	rfid := <-ch

	ver, _ := strconv.ParseUint(r.Form.Get(`ver`), 10, 64)
	offset, _ := strconv.Atoi(r.Form.Get(`offset`))
	rCh := make(chan bool)

	fmt.Println(`new session`, sid, sFile, ver, offset)

	go readLoop(*ws, &rCh)

	bLoop := true
	for {
		select {
		case <-ch:
			bLoop = send(rfid, sFile, &ver, &offset, ws)
		case rc := <-rCh:
			bLoop = rc
		case <-time.After(time.Second * time.Duration(noopInterval)):
			bLoop = sendNoop(ws)
		}
		if !bLoop {
			break
		}
	}

	delete(sessionChan, sid)
	delete(*sessionMap[rfid], sid)
	ws.Close()
}

func readLoop(ws websocket.Conn, rCh *chan bool) {
	for {
		_, _, err := ws.ReadMessage()
		if err == nil {
			select {
			case *rCh <- true:
			default:
			}
		} else {
			*rCh <- false
			return
		}
	}
}

func send(fid uint64, file string, ver *uint64, offset *int, ws *websocket.Conn) bool {

	var tmpVer uint64 = 0
	bReset := false

	fc := filePool[fid]
	if fc == nil {
		fmt.Println(`empty filePool`)
		return true
	}

	fmt.Println(`filePool`, fc)

	f, err := os.Open(file)
	if err != nil {
		return true
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

	if !_send(fid, bReset, tmpVer, buf, ver, offset, ws) {
		return false
	}

	return true
}

func _send(fid uint64, bReset bool, tmpVer uint64, buf *bufio.Reader, ver *uint64, offset *int, ws *websocket.Conn) bool {

	tmpCnt := 0

	prevBuf := []byte(``)

	for {
		sbuf := make([]byte, buffLen)
		n, err := buf.Read(sbuf)

		if err == io.EOF {
			if *offset > 0 && tmpVer > 0 {
				*ver = tmpVer
			}
			break
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
				subs := slen - i
				if subs < 1 {
					break
				}
				tbuf := sbuf[:subs]
				if utf8.Valid(tbuf) {
					pass = true
					prevBuf = sbuf[subs:]
					sbuf = tbuf
					break
				}
			}
			if !pass {
				continue
			}
		}

		sbuf = append([]byte{'>'}, sbuf...)

		tmpCnt += n

		err = wsWrite(ws, sbuf)
		if err != nil {
			fmt.Println(`ws err`, err)
			return false
		}
	}
	return true
}

func sendNoop(ws *websocket.Conn) bool {
	err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait))
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
		file := sessInfo.file

		ftmp := fileMap[file]
		if ftmp > 0 {
			fid = ftmp
		} else {
			fileSerial++
			// fid = atomic.AddUint64(&fileSerial, 1)
			fid = fileSerial
			fileMap[file] = fid
			go scan(fid, file)
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
			*sessionChan[sessInfo.id] <- 1
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
	l := sessionMap[fid]
	if l == nil {
		return
	}

	for sid, _ := range *l {
		go _tend(sid)
	}
}

func _tend(sid uint64) {
	ch := sessionChan[sid]
	if ch == nil {
		return
	}
	select {
	case *ch <- 1:
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

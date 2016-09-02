package main

import "fmt"
import "log"
import "github.com/fsnotify/fsnotify"

var sessionMap = make(map[string]map[int]int)
var sessionChan = make(map[int]chan string)

func main() {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("/tmp/a.txt")
	if err != nil {
		log.Fatal(err)
	}
	<-done

	fmt.Println(`end`)
}

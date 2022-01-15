package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var (
	addr                 = flag.String("addr", "127.0.0.1:12345", "http service address")
	connection           *websocket.Conn
	messageType          int = 1
	watcher              *fsnotify.Watcher
	lastMessageTimestamp int64 = 0
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Resolve cross-domain problems
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func clientConnected(w http.ResponseWriter, r *http.Request) {
	var err error
	connection, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer connection.Close()
	for {
		// var message []byte
		var err error
		messageType, _, err = connection.ReadMessage()
		if err != nil {
			// log.Println("read error:", err)
			break
		}
		// log.Printf("recv: %s", message)
	}
}

func sendReload() {
	if connection != nil {
		if lastMessageTimestamp < time.Now().Unix() {
			err := connection.WriteMessage(messageType, []byte("reload"))
			if err != nil {
				log.Println("write error:", err)
			}
			lastMessageTimestamp = time.Now().Unix()
			log.Println("Sent reload")
		} else {
			log.Println("Too many events this second")
		}
	} else {
		log.Println("No client")
	}
}

func main() {
	log.Println("Starting Watch-to-Websocket")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	cwd = cwd + "/"

	log.Println("Watching path", cwd)
	err = filepath.WalkDir(cwd, func(filename string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// log.Println("Watching", filename)
			return watcher.Add(filename)
		}

		return nil
	})

	if err != nil {
		log.Println("Error", err)
	}

	go func() {
		log.Println("Starting websocket ws://127.0.0.1:12345/ws")
		http.HandleFunc("/ws", clientConnected)
		log.Fatal(http.ListenAndServe(*addr, nil))
	}()

	done := make(chan bool)

	log.Println("Waiting for FS event")

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// log.Println("FS event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					sendReload()
				}

			// watch for errors
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("FS error:", err)
			}
		}
	}()

	<-done

}

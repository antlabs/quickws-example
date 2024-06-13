package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "embed"

	"github.com/antlabs/quickws"
	"github.com/fsnotify/fsnotify"
	"github.com/guonaihong/clop"
)

// key
var conns sync.Map

//go:embed index.html
var indexHTML []byte

var (
	watcher *fsnotify.Watcher
)

type dirWatch struct {
	// 默认是当前目录, 也可以指定目录
	Dir string `clop:"short;long" usage:"watch dir" default:"./"`
}

func main() {
	var err error
	var d dirWatch
	clop.MustBind(&d)
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(d.Dir)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", d.serveWs)
	http.HandleFunc("/file/", d.serveFile)

	go d.watchDir()

	log.Println("Server started at :8080")
	log.Println("websocket started at :8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write(indexHTML)
}

type handler struct {
}

func (h *handler) OnOpen(c *quickws.Conn) {
	conns.Store(c, struct{}{})
}

func (h *handler) OnMessage(c *quickws.Conn, op quickws.Opcode, msg []byte) {
	fmt.Printf("OnMessage: %s, %v\n", msg, op)
	if err := c.WriteTimeout(op, msg, 3*time.Second); err != nil {
		fmt.Println("write fail:", err)
	}
}

func (h *handler) OnClose(c *quickws.Conn, err error) {
	fmt.Println("OnClose: %v", err)
}

func (d *dirWatch) serveWs(w http.ResponseWriter, r *http.Request) {
	c, err := quickws.Upgrade(w, r, quickws.WithServerReplyPing(),
		// quickws.WithServerDecompression(),
		// quickws.WithServerIgnorePong(),
		quickws.WithServerCallback(&handler{}),
		quickws.WithServerReadTimeout(1*time.Hour),
	)
	if err != nil {
		fmt.Println("Upgrade fail:", err)
		return
	}

	c.StartReadLoop()
}

func (d *dirWatch) serveFile(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join(d.Dir, r.URL.Path[len("/file/"):])
	http.ServeFile(w, r, filePath)
}

func (d *dirWatch) watchDir() {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Remove == fsnotify.Remove {
				files, err := os.ReadDir(d.Dir)
				if err != nil {
					log.Println(err)
					return
				}
				fileList := ""
				for _, file := range files {
					fileList += fmt.Sprintf("<li><a href=\"/file/%s\">%s</a></li>", file.Name(), file.Name())
				}
				fileList = "<ul>" + fileList + "</ul>"
				broadcast(fileList)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func broadcast(message string) {
	conns.Range(func(key, value interface{}) bool {
		conn, ok := key.(*quickws.Conn)
		if !ok {
			panic("invalid type")
		}

		err := conn.WriteMessage(quickws.Text, []byte(message))
		if err != nil {
			log.Println("broadcast error:", err)
			conn.Close()
		}
		return true
	})
}

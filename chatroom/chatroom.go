package main

import (
	"fmt"
	"net/http"
	"sync"

	_ "embed"

	"github.com/antlabs/quickws"
)

var (
	clients    = make(map[*quickws.Conn]string)
	clientsMux sync.Mutex
)

//go:embed index.html
var indexHTML []byte

type chatHandler struct{}

func (h *chatHandler) OnOpen(c *quickws.Conn) {
	fmt.Println("New connection")
}

func (h *chatHandler) OnMessage(c *quickws.Conn, op quickws.Opcode, msg []byte) {
	clientsMux.Lock()
	defer clientsMux.Unlock()
	nickname := clients[c]
	message := fmt.Sprintf("%s: %s", nickname, string(msg))
	broadcastMessage(c, message)
}

func (h *chatHandler) OnClose(c *quickws.Conn, err error) {
	fmt.Println("Connection closed:", err)
	clientsMux.Lock()
	defer clientsMux.Unlock()
	delete(clients, c)
	broadcastUserCount()
}

func broadcastMessage(exclude *quickws.Conn, msg string) {
	for client := range clients {
		client.WriteMessage(quickws.Text, []byte(msg))
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write(indexHTML)
}
func broadcastUserCount() {
	message := fmt.Sprintf("Users online: %d", len(clients))
	for client := range clients {
		client.WriteMessage(quickws.Text, []byte(message))
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	nickname := r.URL.Query().Get("nickname")
	if nickname == "" {
		http.Error(w, "Nickname is required", http.StatusBadRequest)
		return
	}

	c, err := quickws.Upgrade(w, r, quickws.WithServerCallback(&chatHandler{}))
	if err != nil {
		fmt.Println("Upgrade fail:", err)
		return
	}

	clientsMux.Lock()
	clients[c] = nickname
	clientsMux.Unlock()

	broadcastUserCount()

	c.StartReadLoop()
}

func main() {
	fmt.Printf("Server started on http://localhost:8080\n")
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", serveHome)
	http.ListenAndServe(":8080", nil)
}

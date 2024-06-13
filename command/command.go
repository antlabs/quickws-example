package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	_ "embed"

	"github.com/antlabs/quickws"
)

//go:embed index.html
var indexHTML []byte

func executeCommand(cmd string) []byte {
	var stdout, stderr bytes.Buffer
	command := exec.Command("sh", "-c", cmd)
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err != nil {
		return []byte(fmt.Sprintf("Error: %s\n%s", err.Error(), stderr.String()))
	}
	return stdout.Bytes()
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write(indexHTML)
}
func main() {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := quickws.Upgrade(w, r, quickws.WithServerOnMessageFunc(func(c *quickws.Conn, op quickws.Opcode, data []byte) {

			log.Printf("Received command: %s", data)
			result := executeCommand(string(data))
			err := c.WriteMessage(quickws.Text, []byte(result))
			if err != nil {
				log.Printf("Write error: %s", err)
				c.Close()
				return
			}
			log.Printf("Sent response: %s", result)
		}))
		if err != nil {
			log.Printf("Upgrade error: %s", err)
			return
		}
		conn.StartReadLoop()
	})

	log.Println("Server started on ws://localhost:8080")
	log.Println("open http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("ListenAndServe error: %s", err)
	}
}

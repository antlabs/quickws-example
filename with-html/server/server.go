package main

import (
	"fmt"
	"net/http"

	"github.com/antlabs/quickws"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.GET("/ws", func(c *gin.Context) {
		con, err := quickws.Upgrade(c.Writer, c.Request, quickws.WithServerCallback(&handler{}))
		if err != nil {
			return
		}
		con.StartReadLoop()
	})
	router.Run()
}

type handler struct{}

func (h *handler) OnOpen(c *quickws.Conn) {
	fmt.Printf("服务端收到一个新的连接")
}

func (h *handler) OnMessage(c *quickws.Conn, op quickws.Opcode, msg []byte) {
	// 如果msg的生命周期不是在OnMessage中结束，需要拷贝一份
	// newMsg := makc([]byte, len(msg))
	// copy(newMsg, msg)

	fmt.Printf("收到客户端消息:%s\n", msg)
	c.WriteMessage(op, msg)
	// os.Stdout.Write(msg)
}

func (h *handler) OnClose(c *quickws.Conn, err error) {
	fmt.Printf("服务端连接关闭:%v\n", err)
}

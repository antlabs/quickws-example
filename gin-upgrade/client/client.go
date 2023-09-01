package main

import (
	"fmt"
	"time"

	"github.com/antlabs/quickws"
	"github.com/antlabs/wsutil/opcode"
)

type handler struct{}

func (h *handler) OnOpen(c *quickws.Conn) {
	fmt.Printf("客户端连接成功\n")
}

func (h *handler) OnMessage(c *quickws.Conn, op quickws.Opcode, msg []byte) {
	// 如果msg的生命周期不是在OnMessage中结束，需要拷贝一份
	// newMsg := makc([]byte, len(msg))
	// copy(newMsg, msg)

	fmt.Printf("收到服务端消息:%s\n", msg)
	c.WriteMessage(op, msg)
	time.Sleep(time.Second)
}

func (h *handler) OnClose(c *quickws.Conn, err error) {
	fmt.Printf("客户端端连接关闭:%v\n", err)
}

func main() {
	c, err := quickws.Dial("ws://127.0.0.1:8080", quickws.WithClientCallback(&handler{}))
	if err != nil {
		fmt.Printf("连接失败:%v\n", err)
		return
	}

	c.WriteMessage(opcode.Text, []byte("hello"))
	c.ReadLoop()
}

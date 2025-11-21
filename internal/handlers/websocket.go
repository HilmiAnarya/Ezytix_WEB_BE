package handlers

import (
	"log"

	"github.com/gofiber/contrib/websocket"
)

func Websocket(c *websocket.Conn) {
	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("WS error:", err)
			break
		}
		c.WriteMessage(mt, []byte("Echo: "+string(msg)))
	}
}

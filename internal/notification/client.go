package notification

import (
	"github.com/gorilla/websocket"
	"log"
)

// Client ใช้สำหรับจัดการการเชื่อมต่อ WebSocket ของแต่ละ client
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		// ส่ง message ผ่าน WebSocket
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}

// NewClient สร้าง client ใหม่
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte),
	}
}

// ฟังก์ชันอ่านข้อความจาก WebSocket
func (c *Client) read() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage() // อ่านข้อความจาก WebSocket
		if err != nil {
			break
		}
		// ใช้ message ต่อไป (เช่น ส่งไปยัง clients อื่นๆ หรือบันทึก)
		log.Printf("Received message: %s", message)

		// ถ้าต้องการ broadcast ข้อความ
		c.hub.broadcast <- message
	}
}

// write ใช้เพื่อส่งข้อความไปยัง WebSocket
func (c *Client) write() {
	defer func() {
		c.conn.Close()
	}()
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}

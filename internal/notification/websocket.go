package notification

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// WSHandler ใช้จัดการการเชื่อมต่อ WebSocket
type WSHandler struct {
	hub *Hub
}

// NewWSHandler สร้าง WSHandler
func NewWSHandler(hub *Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// HandleConnection รับการเชื่อมต่อ WebSocket จาก client
func (h *WSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// เตรียม Header และค่าการเชื่อมต่อ
	headers := http.Header{}
	headers.Set("Origin", r.Header.Get("Origin"))

	// อัพเกรดการเชื่อมต่อจาก HTTP ไปยัง WebSocket
	conn, err := websocket.Upgrade(w, r, headers, 1024, 1024) // เพิ่ม headers, status code และ subprotocols
	if err != nil {
		log.Println("Failed to upgrade WebSocket connection:", err)
		return
	}

	// สร้าง client ใหม่
	client := NewClient(h.hub, conn)

	// ลงทะเบียน client ใหม่ใน hub
	h.hub.register <- client

	// เริ่มฟังก์ชันอ่านและเขียน WebSocket
	go client.read()
	go client.write()
}

package notification

import (
	"encoding/json"
	"log"
	"task-manager/models"
)

// Hub ใช้จัดการ client และ broadcast ข้อความ
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

// NewHub สร้าง Hub ใหม่
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run ฟังก์ชันนี้จะรัน Hub และจัดการการลงทะเบียน client และการ broadcast ข้อความ
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// Broadcast ใช้สำหรับ broadcast ข้อความไปยัง client ทุกตัวที่เชื่อมต่อ
// ในไฟล์ hub.go
func (h *Hub) Broadcast(noti models.Notification) {
	data, err := json.Marshal(noti) // แปลง struct เป็น JSON
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return
	}
	log.Printf("noti: %v", noti)
	h.broadcast <- data
}

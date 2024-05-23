package handlers

import (
	"log"
	"net/http"
	"sync"
	"website/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var (
    clients = make(map[*websocket.Conn]bool)
    chats   = make(map[string]*models.Chat)
    mutex   sync.Mutex
    broadcast = make(chan models.Message)
)

func HandleConnection(conn *websocket.Conn) {
	var newChat models.Chat
	// Generate a unique chat ID and create a new chat session
	newChat.ID = generateUniqueID()
	newChat.ClientID = "client1" // Replace with actual client ID
	newChat.Active = true
	chats[newChat.ID] = &newChat

	chatNotification := struct {
			ClientID string
			ChatID   string
	}{
			ClientID: newChat.ClientID,
			ChatID:   newChat.ID,
	}

	for client := range clients {
			err := client.WriteJSON(chatNotification)
			if err != nil {
					log.Println("Error sending chat notification:", err)
			}
	}

	// Handle messages for the new chat
	clients[conn] = true
	defer func() {
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			conn.Close()
	}()

	for {
			var msg models.Message
			err := conn.ReadJSON(&msg)
			if err != nil {
					log.Println("Error reading message:", err)
					break
			}
			broadcast <- msg
	}
}


func HandleMessages() {
    for {
        msg := <-broadcast
        mutex.Lock()
        for client := range clients {
            err := client.WriteJSON(msg)
            if err != nil {
                log.Println("Error writing message:", err)
                client.Close()
                delete(clients, client)
            }
        }
        mutex.Unlock()
    }
}

func generateUniqueID() string {
    return uuid.New().String()
}


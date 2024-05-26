package handlers

import (
	"context"
	"log"
	"sync"
	"website/database"
	"website/models"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	mutex     sync.Mutex
	broadcast = make(chan models.Message)
)

func HandleConnection(conn *websocket.Conn, username string, chatID string) {
	// Load previous messages from MongoDB for the specific chat room
	loadPreviousMessages(conn, chatID)

	// Register new client
	clients[conn] = true
	defer func() {
		mutex.Lock()
		delete(clients, conn)
		mutex.Unlock()
		conn.Close()
	}()

	// Read messages from the client
	for {
		var msg models.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		// Add the username to the message
		msg.Username = username
		// Save the message to MongoDB
		saveMessage(msg, chatID)
		// Broadcast the message to all clients in the chat room
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
func saveMessage(msg models.Message, chatID string) {
	collection := database.Client.Database("project").Collection("messages" + chatID)
	_, err := collection.InsertOne(context.Background(), msg)
	if err != nil {
		log.Println("Error saving message to MongoDB:", err)
	}
}

func loadPreviousMessages(conn *websocket.Conn, chatID string) {
	collection := database.Client.Database("project").Collection("messages" + chatID)
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Println("Error loading previous messages from MongoDB:", err)
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var msg models.Message
		if err := cursor.Decode(&msg); err != nil {
			log.Println("Error decoding message from MongoDB:", err)
			continue
		}
		err = conn.WriteJSON(msg)
		if err != nil {
			log.Println("Error sending previous message to client:", err)
		}
	}
	if err := cursor.Err(); err != nil {
		log.Println("Cursor error:", err)
	}
}

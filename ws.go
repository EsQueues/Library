package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"website/handlers"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to websocket:", err)
		return
	}
	username := r.URL.Query().Get("username")
	if username == "" {
		username = "Anonymous"
	}
	chatID := r.URL.Query().Get("chatID")
	if chatID == "" {
		chatID = "default"
	}

	handlers.HandleConnection(conn, username, chatID)
}

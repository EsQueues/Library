package main

import (
	"log"
	"net/http"
	"time"
	"website/handlers"
	"website/middleware"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	go handlers.HandleMessages()

	r := mux.NewRouter()

	LoadRoutes(r)

	r.Use(middleware.RateLimitMiddleware)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server listening on port 8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Error starting the server:", err)
	}
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

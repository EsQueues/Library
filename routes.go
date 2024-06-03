package main

import (
	"website/handlers"

	"github.com/gorilla/mux"
)

func LoadRoutes(r *mux.Router) {
	r.HandleFunc("/", handlers.StartHandler).Methods("GET")
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("POST", "GET")
	r.HandleFunc("/confirm", handlers.ConfirmHandler).Methods("GET", "POST")
	r.HandleFunc("/email-confirmed", handlers.EmailConfirmedHandler).Methods("GET")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST", "GET")
	r.HandleFunc("/profile", handlers.ProfileHandler).Methods("GET")
	r.HandleFunc("/delete", handlers.DeleteHandler).Methods("POST")
	r.HandleFunc("/edit", handlers.EditHandler).Methods("GET")
	r.HandleFunc("/update", handlers.UpdateHandler).Methods("POST")

	r.HandleFunc("/filtered-books", handlers.FilterBooksHandler).Methods("GET")
	r.HandleFunc("/buy", handlers.TransactionHandler).Methods("POST")
	r.HandleFunc("/ws", serveWs).Methods("GET")
	r.HandleFunc("/chat-rooms", handlers.ListChatRoomsHandler).Methods("GET")
	r.HandleFunc("/chat", handlers.ChatHandler).Methods("GET")

	r.HandleFunc("/admin", handlers.AdminDashboardHandler).Methods("GET")
	r.HandleFunc("/admin/message", handlers.MessageHandler).Methods("GET", "POST")
	r.HandleFunc("/admin/createChatRoom", handlers.CreateChatRoomHandler).Methods("POST")
	r.HandleFunc("/admin/deleteChatRoom", handlers.DeleteChatRoomHandler).Methods("DELETE")
	r.HandleFunc("/admin/deleteMessage", handlers.DeleteMessageHandler).Methods("POST")
	r.HandleFunc("/admin/chat", handlers.AdminChatHandler).Methods("GET")
	r.HandleFunc("/admin/edit-book", handlers.EditBookHandler).Methods("GET")
	r.HandleFunc("/admin/delete-book", handlers.DeleteBookHandler).Methods("POST")
	r.HandleFunc("/admin/add-book", handlers.AddBookHandler).Methods("POST")
}

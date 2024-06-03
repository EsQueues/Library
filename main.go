package main

import (
	"log"
	"net/http"
	"time"
	"website/middleware"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	LoadRoutes(r)
	r.Use(middleware.RateLimitMiddleware)
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server started on port: http://localhost:8080/")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Error starting the server:", err)
	}
}

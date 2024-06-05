package main

import (
	"context"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"

	"github.com/rs/cors"
)

var (
	database *mongo.Database
)

func main() {
	r := mux.NewRouter()
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Assign the client to your database package
	database = client.Database("project")
	// Set database variable for global use
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080"}, // Replace with your main application's origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)
	r.HandleFunc("/buy", TransactionHandler).Methods("POST")
	server := &http.Server{
		Addr:         ":8081",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Transaction service started on port: http://localhost:8081/")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Error starting the transaction service:", err)
	}
}

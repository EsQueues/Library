package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"strconv"
	"sync"
	"time"
	"website/database"
	"website/models"
)

func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	books, err := getBooksFromDatabase()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("frontend/admin.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
		Books []models.Book
	}{
		Title: "Admin Dashboard",
		Books: books,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	http.ServeFile(w, r, "frontend/admin.html")
}

func getBooksFromDatabase() ([]models.Book, error) {
	client := database.Client
	if client == nil {
		log.Println("MongoDB client is not initialized")
		return nil, errors.New("MongoDB client is not initialized")
	}
	collection := client.Database("project").Collection("books")

	filter := bson.M{}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Println("Error fetching books from MongoDB:", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var books []models.Book
	for cursor.Next(context.Background()) {
		var book models.Book
		err := cursor.Decode(&book)
		if err != nil {
			log.Println("Error decoding book:", err)
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}

func getBookByTitleFromDatabase(title string) (models.Book, error) {
	client := database.Client
	if client == nil {
		log.Println("MongoDB client is not initialized")
		return models.Book{}, errors.New("MongoDB client is not initialized")
	}

	collection := client.Database("project").Collection("books")
	filter := bson.M{"title": title}

	var book models.Book
	err := collection.FindOne(context.Background(), filter).Decode(&book)
	if err != nil {
		log.Println("Error fetching book from MongoDB:", err)
		return models.Book{}, err
	}

	return book, nil
}

func EditBookHandler(w http.ResponseWriter, r *http.Request) {
	// Extract book title from the query parameters
	title := r.URL.Query().Get("title")

	// Fetch book details for editing
	book, err := getBookByTitleFromDatabase(title)
	if err != nil {
		http.Error(w, "Error fetching book details: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the edit book page with pre-filled form values
	tmpl, err := template.ParseFiles("frontend/admin.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, book)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func DeleteBookHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Extract book titles from the form values
	titles := r.Form["title"]

	// Create a channel to receive errors from goroutines
	errCh := make(chan error, len(titles))

	// Iterate over the book titles and delete each book concurrently
	for _, title := range titles {
		go func(title string) {
			err := deleteBookFromDatabase(title)
			if err != nil {
				errCh <- err
				return
			}
			errCh <- nil // Signal successful deletion
		}(title)
	}

	// Wait for all goroutines to finish and collect errors
	for range titles {
		if err := <-errCh; err != nil {
			http.Error(w, "Failed to delete book: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Respond with success message
	http.Redirect(w, r, "/admin", http.StatusSeeOther)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Books deleted successfully"))
}

func AddBookHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
		return
	}

	// Extract book details from the form
	title := r.FormValue("title")
	author := r.FormValue("author")
	genre := r.FormValue("genre")

	// Convert publicationYear to int64
	publicationYearStr := r.FormValue("publicationYear")
	publicationYear, err := strconv.ParseInt(publicationYearStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Publication Year", http.StatusBadRequest)
		return
	}

	// Insert a new record into the database (assuming MongoDB)
	err = addBookToDatabase(title, author, genre, publicationYear)
	if err != nil {
		http.Error(w, "Failed to add book: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to a success page or handle the response as needed
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func updateBookInDatabase(title, author, genre string, publicationYear int64) error {
	collection := database.Client.Database("project").Collection("books")
	filter := bson.M{"title": title}
	update := bson.M{"$set": bson.M{"author": author, "genre": genre, "publicationYear": int32(publicationYear)}}
	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

// Delete a book from the database
func deleteBookFromDatabase(title string) error {

	collection := database.Client.Database("project").Collection("books")
	filter := bson.M{"title": title}
	_, err := collection.DeleteOne(context.Background(), filter)
	return err

}

// Add a new book to the database
func addBookToDatabase(title, author, genre string, publicationYear int64) error {
	collection := database.Client.Database("project").Collection("books")
	book := models.Book{Title: title, Author: author, Genre: genre, PublicationYear: int32(publicationYear)}
	_, err := collection.InsertOne(context.Background(), book)
	return err
}
func MessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		subject := r.FormValue("subject")
		body := r.FormValue("body")

		client := database.Client

		users, err := getUsers(client)
		if err != nil {
			http.Error(w, "Error fetching users", http.StatusInternalServerError)
			return
		}

		var emails []string
		for _, user := range users {
			emails = append(emails, user.Email)
		}

		start := time.Now()

		err = sendMailSimple(subject, body, emails)
		//err = sendMailConcurrent(subject, body, emails, 1)
		//err = sendMailConcurrent(subject, body, emails, 10)
		//err = sendMailConcurrent(subject, body, emails, 100)
		//err = sendMailConcurrent(subject, body, emails, 1000)

		if err != nil {
			http.Error(w, "Error sending email", http.StatusInternalServerError)
			return
		}

		elapsed := time.Since(start)
		fmt.Printf("Sending emails took %s\n", elapsed)

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	} else if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "frontend/admin.html")
	} else {
		// Method not allowed
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func getUsers(client *mongo.Client) ([]models.User, error) {
	var users []models.User

	collection := client.Database("project").Collection("users")

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func sendMailSimple(subject string, body string, to []string) error {
	auth := smtp.PlainAuth(
		"",
		"saiat.kusainov05@gmail.com",
		"mvip fblq yhtq gwqa",
		"smtp.gmail.com")

	msg := "Subject: " + subject + "\n" + body

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		"saiat.kusainov05@gmail.com",
		to,
		[]byte(msg),
	)

	if err != nil {
		return err
	}
	return nil
}

func sendMailConcurrent(subject string, body string, to []string, numWorkers int) error {
	// Define the SMTP server information
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	fromEmail := "saiat.kusainov05@gmail.com"
	password := "your-password" // Replace with your actual password or use an environment variable

	// Set up authentication information.
	auth := smtp.PlainAuth("", fromEmail, password, smtpHost)

	// Prepare the message.
	msg := []byte("Subject: " + subject + "\r\n\r\n" + body)

	// Create a channel to limit concurrent goroutines.
	semaphore := make(chan struct{}, numWorkers)

	// WaitGroup to wait for all goroutines to finish.
	var wg sync.WaitGroup

	// Loop over the recipients and send emails concurrently.
	for _, recipient := range to {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire a worker slot

		go func(recipient string) {
			defer wg.Done()                // Signal that the goroutine is finished.
			defer func() { <-semaphore }() // Release the worker slot.

			// Send the email.
			err := smtp.SendMail(smtpHost+":"+smtpPort, auth, fromEmail, []string{recipient}, msg)
			if err != nil {
				fmt.Printf("Error sending email to %s: %v\n", recipient, err)
				// Handle the error according to your needs.
			}
		}(recipient)
	}

	// Wait for all goroutines to finish.
	wg.Wait()

	return nil
}

func generateID() string {
	return uuid.New().String()
}

func CreateChatRoomHandler(w http.ResponseWriter, r *http.Request) {
	var chat models.Chat
	if err := json.NewDecoder(r.Body).Decode(&chat); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	chat.ID = generateID()
	chat.Active = true

	collection := database.Client.Database("project").Collection("chats")
	_, err := collection.InsertOne(context.Background(), chat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new collection for the chat room's messages
	messageCollectionName := "messages_" + chat.ID
	database.Client.Database("project").CreateCollection(context.Background(), messageCollectionName)

	// Respond with a success message
	response := map[string]interface{}{
		"success": true,
		"message": "Chat room created successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func DeleteChatRoomHandler(w http.ResponseWriter, r *http.Request) {
	chatName := r.URL.Query().Get("name")

	// Delete the chat room
	chatCollection := database.Client.Database("project").Collection("chats")
	filter := bson.M{"id": chatName}
	result, err := chatCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Chat room not found", http.StatusNotFound)
		return
	}

	// Send a success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Chat room deleted successfully"))
}
func ListChatRoomsHandler(w http.ResponseWriter, r *http.Request) {
	client := database.Client
	if client == nil {
		http.Error(w, "MongoDB client is not initialized", http.StatusInternalServerError)
		return
	}

	collection := client.Database("project").Collection("chats")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, "Failed to fetch chat rooms", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var chatRooms []models.Chat
	if err = cursor.All(context.Background(), &chatRooms); err != nil {
		http.Error(w, "Failed to fetch chat rooms", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatRooms)
}

type DeleteMessageRequest struct {
	ChatID    string `json:"chatID"`
	Timestamp string `json:"timestamp"`
}

func DeleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	var request DeleteMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	client := database.Client
	if client == nil {
		http.Error(w, "MongoDB client is not initialized", http.StatusInternalServerError)
		return
	}

	collectionName := "messages_" + request.ChatID
	collection := client.Database("project").Collection(collectionName)

	filter := bson.M{"timestamp": request.Timestamp}
	result, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		http.Error(w, "Failed to delete message", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func AdminChatHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/admin_chat.html")
}

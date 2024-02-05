package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/time/rate"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

var store = sessions.NewCookieStore([]byte("your-secret-key"))

var client *mongo.Client
var limiter = rate.NewLimiter(1, 3)
var log = logrus.New()

type User struct {
	Fullname string `json:"fullname"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Book struct {
	Title           string `json:"title"`
	Author          string `json:"author"`
	Genre           string `json:"genre"`
	PublicationYear int32  `json:"publicationYear"`
	ISBN            string `json:"isbn"`
}

func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	var err error
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}

	// Create a logs folder if it doesn't exist
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		err := os.Mkdir("logs", 0755)
		if err != nil {
			return
		}
	}

	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		println("")
		log.SetOutput(logFile)
		log.SetFormatter(&logrus.TextFormatter{})
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.Error("Failed to log to file, using default stderr")
	}
}
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			// Exceeded request limit
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fullname := r.FormValue("fullname")
		username := r.FormValue("username")
		password := r.FormValue("password")
		user := User{
			Fullname: fullname,
			Username: username,
			Password: password,
		}

		collection := client.Database("project").Collection("users")
		_, err = collection.InsertOne(context.Background(), user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Println("User registered successfully. Redirecting to login page.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	http.ServeFile(w, r, "frontend/register.html")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		collection := client.Database("project").Collection("users")
		filter := bson.M{"username": username}
		var storedUser User
		err = collection.FindOne(context.Background(), filter).Decode(&storedUser)
		if errors.Is(err, mongo.ErrNoDocuments) {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if storedUser.Password != password {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		session, err := store.Get(r, "session-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.Values["username"] = storedUser.Username
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Println("Login successful. Redirecting to profile page.")

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	http.ServeFile(w, r, "frontend/login.html")
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	collection := client.Database("project").Collection("users")
	filter := bson.M{"username": username}
	var storedUser User
	err = collection.FindOne(context.Background(), filter).Decode(&storedUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("frontend/profile.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, storedUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	collection := client.Database("project").Collection("users")
	filter := bson.M{"username": username}
	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
func editHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	collection := client.Database("project").Collection("users")
	filter := bson.M{"username": username}
	var storedUser User
	err = collection.FindOne(context.Background(), filter).Decode(&storedUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("frontend/edit.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, storedUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newFullName := r.FormValue("fullname")
	newUserName := r.FormValue("username")
	newPassword := r.FormValue("password")

	collection := client.Database("project").Collection("users")
	filter := bson.M{"username": username}

	update := bson.M{
		"$set": bson.M{
			"fullname": newFullName,
			"username": newUserName,
			"password": newPassword,
		},
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)

}

const booksPerPage = 10

func filterBooksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Extract page and filter parameters from the request
		pageStr := r.URL.Query().Get("page")
		filterValue := r.URL.Query().Get("filter")

		// Parse page number
		page := 1
		if pageStr != "" {
			p, err := strconv.Atoi(pageStr)
			if err == nil && p > 0 {
				page = p
			}
		}

		// Calculate skip value based on page number
		skip := (page - 1) * booksPerPage

		// Perform sorting and pagination on the server side
		sortBy := r.URL.Query().Get("sort")
		books, err := getSortedAndPaginatedBooks(filterValue, sortBy, skip, booksPerPage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Encode the result as JSON and send it in the response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(books)
	}
}

func getSortedAndPaginatedBooks(filterValue, sortBy string, skip, limit int) ([]Book, error) {
	// Get all books matching the filter
	filter := bson.M{"title": bson.M{"$regex": filterValue, "$options": "i"}}
	books, err := getFilteredBooks(filter)
	if err != nil {
		return nil, err
	}

	// Sort the entire list of books
	books = sortBooks(books, sortBy)

	// Paginate the sorted list
	start := skip
	end := skip + limit
	if start > len(books) {
		start = len(books)
	}
	if end > len(books) {
		end = len(books)
	}

	return books[start:end], nil
}

func getFilteredBooks(filter bson.M) ([]Book, error) {
	collection := client.Database("project").Collection("books")

	indexModel := mongo.IndexModel{
		Keys: bson.M{"title": 1},
	}

	indexOptions := options.CreateIndexes().SetMaxTime(2 * time.Second)

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel, indexOptions)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var books []Book
	for cursor.Next(context.Background()) {
		var book Book
		err := cursor.Decode(&book)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}

func sortBooks(books []Book, sortBy string) []Book {
	switch sortBy {
	case "title":
		sort.Slice(books, func(i, j int) bool {
			return books[i].Title < books[j].Title
		})
	case "author":
		sort.Slice(books, func(i, j int) bool {
			return books[i].Author < books[j].Author
		})
	case "genre":
		sort.Slice(books, func(i, j int) bool {
			return books[i].Genre < books[j].Genre
		})
	case "publicationYear":
		sort.Slice(books, func(i, j int) bool {
			return books[i].PublicationYear < books[j].PublicationYear
		})
	default:
		fmt.Println("Invalid sortBy parameter")
	}

	return books
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received request for /")
		http.ServeFile(w, r, "frontend/index.html")
	})
	r.HandleFunc("/register", rateLimitMiddleware(registerHandler)).Methods("POST", "GET")
	r.HandleFunc("/login", rateLimitMiddleware(loginHandler)).Methods("POST", "GET")
	r.HandleFunc("/profile", rateLimitMiddleware(profileHandler)).Methods("GET")
	r.HandleFunc("/delete", rateLimitMiddleware(deleteHandler)).Methods("POST")
	r.HandleFunc("/edit", rateLimitMiddleware(editHandler)).Methods("GET")
	r.HandleFunc("/update", rateLimitMiddleware(updateHandler)).Methods("POST")
	r.HandleFunc("/filtered-books", rateLimitMiddleware(filterBooksHandler)).Methods("GET")

	fmt.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", r)

	if err != nil {
		log.Fatal("Error starting the server:", err)
		return
	}
}

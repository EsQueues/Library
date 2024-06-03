package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"sort"
	"strconv"
	"time"
	"website/database" // Import the database package
	"website/models"
)

var client *mongo.Client

func init() {
	// Initialize the MongoDB client when the package is initialized
	if err := database.InitMongoDB(); err != nil {
		panic(err)
	}
	client = database.Client
}

const booksPerPage = 10

func FilterBooksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		pageStr := r.URL.Query().Get("page")
		filterValue := r.URL.Query().Get("filter")

		page := 1
		if pageStr != "" {
			p, err := strconv.Atoi(pageStr)
			if err == nil && p > 0 {
				page = p
			}
		}

		skip := (page - 1) * booksPerPage

		sortBy := r.URL.Query().Get("sort")
		books, err := getSortedAndPaginatedBooks(filterValue, sortBy, skip, booksPerPage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(books)
	}
}

func getSortedAndPaginatedBooks(filterValue, sortBy string, skip, limit int) ([]models.Book, error) {
	filter := bson.M{"title": bson.M{"$regex": filterValue, "$options": "i"}}
	books, err := getFilteredBooks(filter)
	if err != nil {
		return nil, err
	}

	books = sortBooks(books, sortBy)

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

func getFilteredBooks(filter bson.M) ([]models.Book, error) {
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

	var books []models.Book
	for cursor.Next(context.Background()) {
		var book models.Book
		err := cursor.Decode(&book)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}

func sortBooks(books []models.Book, sortBy string) []models.Book {
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

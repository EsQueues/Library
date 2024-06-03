package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"website/models"

	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var store = sessions.NewCookieStore([]byte("your-secret-key"))

// EmailAuth contains the SMTP authentication details
type EmailAuth struct {
	Username string
	Password string
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Replace this with your actual logic to get the username from the session
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok {
		// If the username is not found in the session, set it to an empty string or handle the case accordingly
		username = ""
	}

	// Parse the HTML templates
	tmpl := template.Must(template.ParseFiles("index.html"))

	// Execute the templates with the username data
	err = tmpl.Execute(w, struct{ Username string }{Username: username})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// SendMail sends an email with the given subject, body, and recipient
func SendMail(subject string, body string, to []string, auth EmailAuth) error {
	authString := smtp.PlainAuth("", auth.Username, auth.Password, "smtp.gmail.com")

	msg := "From: " + auth.Username + "\r\n" +
		"To: " + to[0] + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587", authString, auth.Username, to, []byte(msg))
	if err != nil {
		return err
	}
	return nil
}

const CodeLength = 6 // Adjust the code length as needed

func GenerateCode() string {
	const charset = "0123456789"
	randomCode := make([]byte, CodeLength)
	for i := range randomCode {
		randomCode[i] = charset[rand.Intn(len(charset))]
	}
	return string(randomCode)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fullname := r.FormValue("fullname")
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		code := GenerateCode() // Generate a unique code for email confirmation

		user := models.User{
			Fullname: fullname,
			Username: username,
			Email:    email,
			Password: password,
			Code:     code, // Store the code with the user
		}

		collection := client.Database("project").Collection("users")
		_, err = collection.InsertOne(context.Background(), user)
		if err != nil {
			log.Println("Error inserting user into database:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send confirmation email with the code
		subject := "Confirmation Code"
		body := fmt.Sprintf("Hello %s, your confirmation code is: %s", username, code)
		err = SendMail(subject, body, []string{email}, EmailAuth{"saiat.kusainov05@gmail.com", "mvip fblq yhtq gwqa"})
		if err != nil {
			log.Println("Error sending confirmation email:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect to a page indicating that the user needs to confirm their email
		http.Redirect(w, r, "/confirm", http.StatusSeeOther)
		return
	}

	// Serve the registration page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	http.ServeFile(w, r, "frontend/register.html")
}

func ConfirmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		code := r.FormValue("code")
		if code == "" {
			http.Error(w, "Code is required", http.StatusBadRequest)
			return
		}

		// Check if the code matches the one stored in the database
		collection := client.Database("project").Collection("users")
		filter := bson.M{"code": code}
		var user models.User
		err = collection.FindOne(context.Background(), filter).Decode(&user)
		if err != nil {
			http.Error(w, "Invalid code", http.StatusBadRequest)
			return
		}
		//
		update := bson.M{"$unset": bson.M{"code": ""}, "$set": bson.M{"active": true}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect to a page indicating successful email confirmation
		http.Redirect(w, r, "/email-confirmed", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	http.ServeFile(w, r, "frontend/confirm.html")
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == "admin" && password == "12345" {
			// Redirect to the admin page
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

		collection := client.Database("project").Collection("users")
		filter := bson.M{"username": username, "active": true}
		var storedUser models.User
		err = collection.FindOne(context.Background(), filter).Decode(&storedUser)
		if errors.Is(err, mongo.ErrNoDocuments) {
			http.Redirect(w, r, "/login?error=Invalid username or password", http.StatusSeeOther)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if storedUser.Password != password {
			http.Redirect(w, r, "/login?error=Invalid username or password", http.StatusSeeOther)
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

		// Redirect to the profile page after successful login
		http.Redirect(w, r, "/?username="+storedUser.Username, http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	http.ServeFile(w, r, "frontend/login.html")
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
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
	var storedUser models.User
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

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
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
func EditHandler(w http.ResponseWriter, r *http.Request) {
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
	var storedUser models.User
	err = collection.FindOne(context.Background(), filter).Decode(&storedUser)
	if err == mongo.ErrNoDocuments {
		// No user found with the provided username
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		// Other errors encountered while querying the database
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		// Parse form values
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Update user with new full name
		fullName := r.FormValue("fullname")

		update := bson.M{"$set": bson.M{"fullname": fullName}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect to profile page after successful update
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// Render edit page with user data
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

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
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
	newPassword := r.FormValue("password")

	collection := client.Database("project").Collection("users")

	filter := bson.M{"username": username} // Filter by the existing username

	update := bson.M{
		"$set": bson.M{
			"fullname": newFullName,
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
func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var requestData struct {
		BookID string `json:"bookId"`
	}
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bookID, err := primitive.ObjectIDFromHex(requestData.BookID)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	collection := client.Database("project").Collection("users")
	filter := bson.M{"username": username}
	update := bson.M{"$addToSet": bson.M{"cart": bookID}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetCartHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	collection := client.Database("project").Collection("users")
	filter := bson.M{"username": username}
	var user models.User
	err = collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	booksCollection := client.Database("project").Collection("books")
	var books []models.Book
	cursor, err := booksCollection.Find(context.Background(), bson.M{"_id": bson.M{"$in": user.Cart}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var book models.Book
		err := cursor.Decode(&book)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

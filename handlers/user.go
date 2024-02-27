package handlers

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"website/models"

	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var store = sessions.NewCookieStore([]byte("your-secret-key"))

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
		user := models.User{
			Fullname: fullname,
			Username: username,
			Email:    email,
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

func LoginHandler(w http.ResponseWriter, r *http.Request) {
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
		var storedUser models.User
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

package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Chat struct {
	ID       string
	Name     string
	AdminID  string
	Messages []Message
	Active   bool
}

type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ChatID    string             `bson:"chat_id"`
	SenderID  string             `bson:"sender_id"`
	Username  string             `bson:"username"`
	Content   string             `bson:"content"`
	Timestamp time.Time          `bson:"timestamp"`
}
type Book struct {
	Title           string `json:"title"`
	Author          string `json:"author"`
	Genre           string `json:"genre"`
	PublicationYear int32  `json:"publicationYear"`
	ISBN            string `json:"isbn"`
	Price           int    `json:"price"` // Add price field for book
}

type Transaction struct {
	Name        string   `json:"name"`
	Surname     string   `json:"surname"`
	Email       string   `json:"email"`
	PhoneNumber string   `json:"phoneNumber"`
	Books       []string `json:"books"`
}

type User struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Receive  chan []byte
	Fullname string               `json:"fullname"`
	Username string               `json:"username"`
	Password string               `json:"password"`
	Email    string               `json:"email"`
	Code     string               `json:"code"`
	Active   bool                 `json:"active"` // The active status of the user
	Cart     []primitive.ObjectID `json:"cart,omitempty" bson:"cart,omitempty"`
}

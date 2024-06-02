package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	BookList []string             `json:"booklist,omitempty" bson:"booklist,omitempty"` // Add booklist field for user's purchased books
}

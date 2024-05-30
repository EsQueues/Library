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

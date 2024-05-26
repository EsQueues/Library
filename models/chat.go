package models

import "time"

type Chat struct {
	ID       string
	Name     string
	AdminID  string
	Messages []Message
	Active   bool
}

type Message struct {
	SenderID  string
	Username  string `json:"Username"`
	Content   string
	Timestamp time.Time
}

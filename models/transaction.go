package models

type Transaction struct {
	Name        string   `json:"name"`
	Surname     string   `json:"surname"`
	Email       string   `json:"email"`
	PhoneNumber string   `json:"phoneNumber"`
	Books       []string `json:"books"`
}

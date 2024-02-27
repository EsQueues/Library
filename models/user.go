package models

type User struct {
	Fullname string `json:"fullname"`
	Username string `json:"username"`
	Password string `json:"password"`
}

package models

type Book struct {
	Title           string `json:"title"`
	Author          string `json:"author"`
	Genre           string `json:"genre"`
	PublicationYear int32  `json:"publicationYear"`
	ISBN            string `json:"isbn"`
	Price           int    `json:"price"` // Add price field for book
}

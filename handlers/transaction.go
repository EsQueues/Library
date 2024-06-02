package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/jung-kurt/gofpdf"
	"gopkg.in/gomail.v2"
	"io"
	"log"
	"net/http"
	"time"
	"website/database"
	"website/models"

	"golang.org/x/net/context"
)

// PaymentRequest combines payment information and book names
type PaymentRequest struct {
	Transaction models.Transaction `json:"transaction"`
	BookNames   []string           `json:"bookNames"`
}

// TransactionHandler handles the request to save a transaction to MongoDB and send an email
func TransactionHandler(w http.ResponseWriter, r *http.Request) {
	// Parse payment request from request body
	var paymentRequest PaymentRequest
	err := json.NewDecoder(r.Body).Decode(&paymentRequest)
	if err != nil {
		log.Printf("Failed to parse payment information: %v", err)
		http.Error(w, `{"message": "Failed to parse payment information"}`, http.StatusBadRequest)
		return
	}

	// Save the transaction to MongoDB
	err = saveTransactionToDatabase(paymentRequest.Transaction, paymentRequest.BookNames)
	if err != nil {
		log.Printf("Failed to save transaction to database: %v", err)
		http.Error(w, `{"message": "Failed to save transaction to database"}`, http.StatusInternalServerError)
		return
	}

	// Generate PDF from transaction data
	pdfBytes, err := generatePDF(paymentRequest.Transaction, paymentRequest.BookNames)
	if err != nil {
		log.Printf("Failed to generate PDF: %v", err)
		http.Error(w, `{"message": "Failed to generate PDF"}`, http.StatusInternalServerError)
		return
	}

	// Read form data
	to := paymentRequest.Transaction.Email
	subject := "Transaction Bill"
	body := "Please find attached transaction bill."

	// Send the email with the PDF attachment
	err = sendEmail(to, subject, body, pdfBytes)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		http.Error(w, `{"message": "Failed to send email"}`, http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Transaction saved successfully and email sent"}`))
}

// saveTransactionToDatabase saves the transaction to the MongoDB collection
func saveTransactionToDatabase(transaction models.Transaction, bookNames []string) error {
	// Combine payment information and book names into a single transaction obj
	transaction.Books = bookNames

	collection := database.Client.Database("project").Collection("transactions")
	_, err := collection.InsertOne(context.Background(), transaction)
	return err
}
func generatePDF(transaction models.Transaction, books []string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "IP KUSAIN SAYAT") // Company name
	pdf.Cell(40, 10, "LIBRARY")
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Transaction Bill")
	pdf.Ln(10)

	// Current date and time
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	pdf.Cell(40, 10, "Date: "+currentTime)
	pdf.Ln(10)

	// Transaction details
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Name: "+transaction.Name)
	pdf.Ln(8)
	pdf.Cell(40, 10, "Surname: "+transaction.Surname)
	pdf.Ln(8)
	pdf.Cell(40, 10, "Email: "+transaction.Email)
	pdf.Ln(8)
	pdf.Cell(40, 10, "Phone Number: "+transaction.PhoneNumber)
	pdf.Ln(10)

	// Book details
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Books:")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 12)
	for _, book := range books {
		pdf.Cell(40, 10, "- "+book)
		pdf.Ln(8)
	}
	pdf.Cell(40, 10, "Come back again!!!")

	var buffer bytes.Buffer
	err := pdf.Output(&buffer)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func sendEmail(to string, subject string, body string, attachment []byte) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "saiat.kusainov05@gmail.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	m.Attach("transaction.pdf", gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(attachment)
		return err
	}))

	d := gomail.NewDialer("smtp.gmail.com", 587, "saiat.kusainov05@gmail.com", "mvip fblq yhtq gwqa")

	return d.DialAndSend(m)
}

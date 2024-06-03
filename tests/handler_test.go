package tests

import (
	"bytes"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"net/http/httptest"
	"testing"
	"unicode"
	"website/handlers"
	"website/models"
)

func TestIfGeneratedCodeIsCorrect(t *testing.T) {
	code := handlers.GenerateCode()
	if len(code) != handlers.CodeLength {
		t.Errorf("generated code length is incorrect: got %d, want %d", len(code), handlers.CodeLength)
	}
	for _, char := range code {
		if !unicode.IsDigit(char) {
			t.Errorf("generated code contains non-digit characters: %s", code)
		}
	}
}

func TestIfUserAfterRegistrationSavedInDBCorrectly(t *testing.T) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
		}
	}(client, context.Background())

	collection := client.Database("project").Collection("users")

	registerData := []byte(`fullname=Test User&username=testuser&email=test@example.com&password=12345`)
	registerReq, err := http.NewRequest("POST", "/register", bytes.NewBuffer(registerData))
	if err != nil {
		t.Fatal(err)
	}
	registerReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	registerResp := httptest.NewRecorder()
	handlers.RegisterHandler(registerResp, registerReq)

	if registerResp.Code != http.StatusSeeOther {
		t.Fatalf("Expected status code %d, got %d", http.StatusSeeOther, registerResp.Code)
	}

	filter := bson.M{"username": "testuser"}
	var user models.User
	err = collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		t.Fatalf("Failed to find user in database: %v", err)
	}

	if user.Fullname != "Test User" || user.Email != "test@example.com" {
		t.Errorf("Unexpected user data in database")
	}
}

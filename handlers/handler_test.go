package handlers

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
	"website/models"
)

func TestIfGeneratedCodeIsCorrect(t *testing.T) {
	code := GenerateCode()
	if len(code) != codeLength {
		t.Errorf("generated code length is incorrect: got %d, want %d", len(code), codeLength)
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
	RegisterHandler(registerResp, registerReq)

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

func TestUserLifecycle(t *testing.T) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			t.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}(client, context.Background())

	collection := client.Database("project").Collection("users")
	filter := bson.M{"username": "testuser"}
	update := bson.M{"$set": bson.M{"active": true}}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		t.Fatalf("Failed to update user in database: %v", err)
	}

	loginData := []byte(`username=testuser&password=12345`)
	loginReq, err := http.NewRequest("POST", "/login", bytes.NewBuffer(loginData))
	if err != nil {
		t.Fatal(err)
	}
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginResp := httptest.NewRecorder()
	LoginHandler(loginResp, loginReq)

	if loginResp.Code != http.StatusSeeOther {
		t.Fatalf("Expected status code %d, got %d", http.StatusSeeOther, loginResp.Code)
	}

	deleteReq, err := http.NewRequest("POST", "/delete", nil)
	if err != nil {
		t.Fatal(err)
	}
	deleteResp := httptest.NewRecorder()
	DeleteHandler(deleteResp, deleteReq)

	if deleteResp.Code != http.StatusSeeOther {
		t.Fatalf("Expected status code %d, got %d", http.StatusSeeOther, deleteResp.Code)
	}
}

package handlers

import (
	"context"

	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestLoginRedirect(t *testing.T) {
	// Create a new context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a new Chrome instance
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	// Navigate to the login page
	if err := chromedp.Run(ctx, chromedp.Navigate("http://localhost:8080/login")); err != nil {
		t.Fatalf("Failed to navigate to the login page: %v", err)
	}

	// Find and fill the username field
	if err := chromedp.Run(ctx, chromedp.SendKeys("#username", "admin")); err != nil {
		t.Fatalf("Failed to enter username: %v", err)
	}

	// Find and fill the password field
	if err := chromedp.Run(ctx, chromedp.SendKeys("#password", "12345")); err != nil {
		t.Fatalf("Failed to enter password: %v", err)
	}

	// Find and click the submit button
	if err := chromedp.Run(ctx, chromedp.Click("button[type='submit']")); err != nil {
		t.Fatalf("Failed to click the submit button: %v", err)
	}

	// Wait for the redirect
	time.Sleep(2 * time.Second)

	// Get the current URL and verify the redirection
	var currentURL string
	if err := chromedp.Run(ctx, chromedp.Location(&currentURL)); err != nil {
		t.Fatalf("Failed to get the current URL: %v", err)
	}

	expectedURL := "http://localhost:8080/admin"
	if currentURL != expectedURL {
		t.Errorf("Login failed! Expected /admin but got: %s", currentURL)
	}
}
func TestNormalUserLoginRedirect(t *testing.T) {
	// Create a new context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a new Chrome instance
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	// Navigate to the login page
	if err := chromedp.Run(ctx, chromedp.Navigate("http://localhost:8080/login")); err != nil {
		t.Fatalf("Failed to navigate to the login page: %v", err)
	}

	// Find and fill the username field
	if err := chromedp.Run(ctx, chromedp.SendKeys("#username", "nomad12")); err != nil {
		t.Fatalf("Failed to enter username: %v", err)
	}

	// Find and fill the password field
	if err := chromedp.Run(ctx, chromedp.SendKeys("#password", "qwert")); err != nil {
		t.Fatalf("Failed to enter password: %v", err)
	}

	// Find and click the submit button
	if err := chromedp.Run(ctx, chromedp.Click("button[type='submit']")); err != nil {
		t.Fatalf("Failed to click the submit button: %v", err)
	}

	// Wait for the redirect
	time.Sleep(2 * time.Second)

	// Get the current URL and verify the redirection
	var currentURL string
	if err := chromedp.Run(ctx, chromedp.Location(&currentURL)); err != nil {
		t.Fatalf("Failed to get the current URL: %v", err)
	}

	expectedURL := "http://localhost:8080/"
	if currentURL != expectedURL {
		t.Errorf("Login failed! Expected / but got: %s", currentURL)
	}
}

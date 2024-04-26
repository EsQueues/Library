package handlers

import (
	"context"

	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestLoginRedirect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	if err := chromedp.Run(ctx, chromedp.Navigate("http://localhost:8080/login")); err != nil {
		t.Fatalf("Failed to navigate to the login page: %v", err)
	}

	if err := chromedp.Run(ctx, chromedp.SendKeys("#username", "admin")); err != nil {
		t.Fatalf("Failed to enter username: %v", err)
	}

	if err := chromedp.Run(ctx, chromedp.SendKeys("#password", "12345")); err != nil {
		t.Fatalf("Failed to enter password: %v", err)
	}

	if err := chromedp.Run(ctx, chromedp.Click("button[type='submit']")); err != nil {
		t.Fatalf("Failed to click the submit button: %v", err)
	}

	time.Sleep(2 * time.Second)

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	if err := chromedp.Run(ctx, chromedp.Navigate("http://localhost:8080/login")); err != nil {
		t.Fatalf("Failed to navigate to the login page: %v", err)
	}

	if err := chromedp.Run(ctx, chromedp.SendKeys("#username", "nomad12")); err != nil {
		t.Fatalf("Failed to enter username: %v", err)
	}

	if err := chromedp.Run(ctx, chromedp.SendKeys("#password", "qwert")); err != nil {
		t.Fatalf("Failed to enter password: %v", err)
	}

	if err := chromedp.Run(ctx, chromedp.Click("button[type='submit']")); err != nil {
		t.Fatalf("Failed to click the submit button: %v", err)
	}

	time.Sleep(2 * time.Second)

	var currentURL string
	if err := chromedp.Run(ctx, chromedp.Location(&currentURL)); err != nil {
		t.Fatalf("Failed to get the current URL: %v", err)
	}

	expectedURL := "http://localhost:8080/"
	if currentURL != expectedURL {
		t.Errorf("Login failed! Expected / but got: %s", currentURL)
	}
}

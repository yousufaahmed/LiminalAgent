package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

const (
	processURL = "https://api.tabscanner.com/api/2/process"
	resultBase = "https://api.tabscanner.com/api/result/"
)

type processResponse struct {
	Token      string `json:"token"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Code       int    `json:"code"`
}

type apiError struct {
	Message    string `json:"message"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Success    bool   `json:"success"`
	Code       int    `json:"code"`
}

func main() {
	// Load .env file
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s /path/to/receipt.jpg\n", os.Args[0])
		os.Exit(2)
	}

	imagePath := os.Args[1]
	apiKey := os.Getenv("TABSCANNER_APIKEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Missing env var TABSCANNER_APIKEY")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	token, err := callProcess(ctx, apiKey, imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "process error: %v\n", err)
		os.Exit(1)
	}

	// Tabscanner recommends waiting ~5s before polling. :contentReference[oaicite:3]{index=3}
	time.Sleep(5 * time.Second)

	resultJSON, err := pollResult(ctx, apiKey, token, 1*time.Second, 30)
	if err != nil {
		fmt.Fprintf(os.Stderr, "result error: %v\n", err)
		os.Exit(1)
	}

	// Pretty print
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, resultJSON, "", "  "); err != nil {
		// If it's not valid JSON for some reason, just print raw.
		fmt.Println(string(resultJSON))
		return
	}
	fmt.Println(pretty.String())
}

// callProcess uploads the image to /api/2/process and returns the token.
func callProcess(ctx context.Context, apiKey, imagePath string) (string, error) {
	f, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// "file" is required. :contentReference[oaicite:4]{index=4}
	part, err := writer.CreateFormFile("file", filepath.Base(imagePath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, f); err != nil {
		return "", err
	}

	// Optional form fields (examples):
	// documentType: receipt|invoice|auto (default receipt) :contentReference[oaicite:5]{index=5}
	_ = writer.WriteField("documentType", "receipt")
	// region: ISO 2-alpha country code e.g. gb for United Kingdom :contentReference[oaicite:6]{index=6}
	_ = writer.WriteField("region", "gb")
	// defaultDateParsing: m/d or d/m (optional) :contentReference[oaicite:7]{index=7}
	_ = writer.WriteField("defaultDateParsing", "d/m")

	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, processURL, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// API key is passed via header named "apikey". :contentReference[oaicite:8]{index=8}
	req.Header.Set("apikey", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)

	// Try parse as success response
	var pr processResponse
	if err := json.Unmarshal(respBytes, &pr); err == nil && pr.Token != "" {
		return pr.Token, nil
	}

	// Otherwise parse error object if present
	var ae apiError
	if err := json.Unmarshal(respBytes, &ae); err == nil && ae.Message != "" {
		return "", fmt.Errorf("tabscanner error (status_code=%d code=%d): %s", ae.StatusCode, ae.Code, ae.Message)
	}

	return "", fmt.Errorf("unexpected response (http=%d): %s", resp.StatusCode, string(respBytes))
}

// pollResult polls /api/result/{token} until result available or attempts exhausted.
// Tabscanner returns status_code 202 when result is available, 301 when not yet. :contentReference[oaicite:9]{index=9}
func pollResult(ctx context.Context, apiKey, token string, interval time.Duration, maxAttempts int) ([]byte, error) {
	if token == "" {
		return nil, errors.New("empty token")
	}

	url := resultBase + token

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("apikey", apiKey) // :contentReference[oaicite:10]{index=10}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Peek status_code if present
		var meta struct {
			StatusCode int    `json:"status_code"`
			Status     string `json:"status"`
			Code       int    `json:"code"`
			Message    string `json:"message"`
		}
		_ = json.Unmarshal(b, &meta)

		// TabScanner returns status_code=3 with status="done" when result is ready
		if meta.StatusCode == 3 && meta.Status == "done" {
			return b, nil
		}
		// status_code=2 means still processing
		if meta.StatusCode == 2 {
			time.Sleep(interval)
			continue
		}

		// Any other code: treat as error and surface payload
		return nil, fmt.Errorf("unexpected result status_code=%d attempt=%d/%d payload=%s", meta.StatusCode, attempt, maxAttempts, string(b))
	}

	return nil, fmt.Errorf("result not available after %d attempts", maxAttempts)
}

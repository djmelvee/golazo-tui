package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// reJWT matches a JWT — three base64url segments separated by dots.
var reJWT = regexp.MustCompile(`eyJ[A-Za-z0-9_\-]+=*\.[A-Za-z0-9_\-]+=*\.[A-Za-z0-9_\-]+=*`)

// Register creates a new account on the worldcup26.ir API and returns the JWT.
// It handles any response format: plain JWT string, JSON flat, or JSON nested.
// The full raw response is included in errors so callers can debug failures.
func Register(ctx context.Context, base string) (string, error) {
	hostname, _ := os.Hostname()
	hostname = sanitize(hostname)
	if hostname == "" {
		hostname = "golazo"
	}
	suffix := time.Now().Unix() % 1000000
	username := fmt.Sprintf("golazo-%s-%06d", hostname, suffix)
	if len(username) > 30 {
		username = username[len(username)-30:]
	}
	password := username + "-pw"
	email := username + "@example.com"

	body, err := json.Marshal(registerRequest{
		Username: username,
		Password: password,
		Email:    email,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/auth/register", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	hc := &http.Client{Timeout: 10 * time.Second}
	resp, err := hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("register: %w", err)
	}
	defer resp.Body.Close()

	// Read full body for flexible parsing and error messages.
	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("register: read body: %w", err)
	}
	rawBody := strings.TrimSpace(string(rawBytes))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("register: HTTP %d — %s", resp.StatusCode, truncate(rawBody, 120))
	}

	// Case 1: body is a plain JWT (no JSON envelope).
	if m := reJWT.FindString(rawBody); m != "" {
		return m, nil
	}

	// Case 2: JSON object — search any field at any nesting depth.
	var obj map[string]json.RawMessage
	if json.Unmarshal(rawBytes, &obj) == nil {
		if t := extractJWT(obj); t != "" {
			return t, nil
		}
	}

	// Case 3: JSON string value (the token itself, quoted).
	var s string
	if json.Unmarshal(rawBytes, &s) == nil {
		if m := reJWT.FindString(s); m != "" {
			return m, nil
		}
	}

	return "", fmt.Errorf("register: no JWT in response — %s", truncate(rawBody, 160))
}

// extractJWT recursively searches a JSON object for any JWT-shaped string value.
func extractJWT(obj map[string]json.RawMessage) string {
	for _, v := range obj {
		var s string
		if json.Unmarshal(v, &s) == nil {
			if m := reJWT.FindString(s); m != "" {
				return m
			}
			continue
		}
		var nested map[string]json.RawMessage
		if json.Unmarshal(v, &nested) == nil {
			if t := extractJWT(nested); t != "" {
				return t
			}
		}
	}
	return ""
}

var reSanitize = regexp.MustCompile(`[^a-z0-9-]`)

func sanitize(s string) string {
	return reSanitize.ReplaceAllString(strings.ToLower(s), "")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

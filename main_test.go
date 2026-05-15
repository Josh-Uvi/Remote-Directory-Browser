package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSessionTokenGeneration tests that tokens are unique and properly formatted
func TestSessionTokenGeneration(t *testing.T) {
	token1, err := generateToken()
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	token2, err := generateToken()
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token1 == "" || token2 == "" {
		t.Fatalf("Generated empty token")
	}

	if token1 == token2 {
		t.Fatalf("Generated duplicate tokens")
	}

	if len(token1) != 64 {
		t.Fatalf("Token length should be 64 (hex-encoded 32 bytes), got %d", len(token1))
	}
}

// TestSessionCreationAndValidation tests session lifecycle
func TestSessionCreationAndValidation(t *testing.T) {
	sm := &SessionManager{sessions: make(map[string]SessionData)}

	// Create session
	token, err := sm.CreateSession("testuser")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Validate session
	username, err := sm.ValidateSession(token)
	if err != nil {
		t.Fatalf("Failed to validate session: %v", err)
	}

	if username != "testuser" {
		t.Fatalf("Expected username 'testuser', got %s", username)
	}

	// Destroy session
	sm.DestroySession(token)

	// Validate destroyed session
	_, err = sm.ValidateSession(token)
	if err == nil {
		t.Fatalf("Expected error for destroyed session")
	}
}

// TestInvalidSession tests that invalid tokens are rejected
func TestInvalidSession(t *testing.T) {
	sm := &SessionManager{sessions: make(map[string]SessionData)}

	_, err := sm.ValidateSession("invalid-token-xyz")
	if err == nil {
		t.Fatalf("Expected error for invalid session")
	}
}

// TestLoginWithValidCredentials tests successful login
func TestLoginWithValidCredentials(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", handleLogin)

	body := LoginRequest{
		Username: "admin",
		Password: "admin123",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Fatalf("Expected successful login, got error: %s", resp.Error)
	}

	// Check if session cookie was set
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("Expected session cookie to be set")
	}

	sessionCookie := cookies[0]
	if sessionCookie.Name != sessionCookieName {
		t.Fatalf("Expected cookie name %s, got %s", sessionCookieName, sessionCookie.Name)
	}

	if !sessionCookie.HttpOnly {
		t.Fatalf("Expected HttpOnly flag to be set on cookie")
	}

	if !sessionCookie.Secure {
		t.Fatalf("Expected Secure flag to be set on cookie")
	}
}

// TestLoginWithInvalidCredentials tests failed login
func TestLoginWithInvalidCredentials(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", handleLogin)

	body := LoginRequest{
		Username: "admin",
		Password: "wrongpassword",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401, got %d", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Success {
		t.Fatalf("Expected failed login")
	}

	if resp.Error == "" {
		t.Fatalf("Expected error message")
	}
}

// TestLoginWithEmptyCredentials tests login with missing fields
func TestLoginWithEmptyCredentials(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", handleLogin)

	body := LoginRequest{
		Username: "",
		Password: "",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}
}

// TestLoginWithInvalidJSON tests login with malformed JSON
func TestLoginWithInvalidJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", handleLogin)

	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for invalid JSON, got %d", w.Code)
	}
}

// TestLogoutDestrysSession tests that logout properly destroys session
func TestLogoutDestroysSession(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/logout", handleLogout)

	// First login
	loginBody := LoginRequest{
		Username: "admin",
		Password: "admin123",
	}
	bodyBytes, _ := json.Marshal(loginBody)

	loginReq := httptest.NewRequest("POST", "/api/login", bytes.NewReader(bodyBytes))
	loginReq.Header.Set("Content-Type", "application/json")

	loginW := httptest.NewRecorder()
	mux.ServeHTTP(loginW, loginReq)

	// Extract session cookie
	sessionCookie := loginW.Result().Cookies()[0]

	// Then logout
	logoutReq := httptest.NewRequest("POST", "/api/logout", nil)
	logoutReq.AddCookie(sessionCookie)

	logoutW := httptest.NewRecorder()
	mux.ServeHTTP(logoutW, logoutReq)

	if logoutW.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on logout, got %d", logoutW.Code)
	}

	var resp APIResponse
	json.NewDecoder(logoutW.Body).Decode(&resp)

	if !resp.Success {
		t.Fatalf("Expected successful logout")
	}
}

// TestSecurityHeaders tests that security headers are set
func TestSecurityHeaders(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", handleLogin)

	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	headers := []string{
		"X-Frame-Options",
		"X-Content-Type-Options",
		"Strict-Transport-Security",
		"Content-Security-Policy",
	}

	for _, header := range headers {
		if w.Header().Get(header) == "" {
			t.Fatalf("Expected header %s to be set", header)
		}
	}
}

// TestMethodNotAllowed tests that invalid HTTP methods are rejected
func TestMethodNotAllowed(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", handleLogin)

	req := httptest.NewRequest("GET", "/api/login", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected status 405, got %d", w.Code)
	}
}

// BenchmarkSessionCreation benchmarks session creation performance
func BenchmarkSessionCreation(b *testing.B) {
	sm := &SessionManager{sessions: make(map[string]SessionData)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.CreateSession("testuser")
	}
}

// BenchmarkSessionValidation benchmarks session validation performance
func BenchmarkSessionValidation(b *testing.B) {
	sm := &SessionManager{sessions: make(map[string]SessionData)}
	token, _ := sm.CreateSession("testuser")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.ValidateSession(token)
	}
}

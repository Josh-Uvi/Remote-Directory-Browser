package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SessionData holds session information
type SessionData struct {
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionManager manages sessions in-memory
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]SessionData
}

// LoginRequest represents login API request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// APIResponse represents a generic API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// DirEntry represents a file or directory entry
type DirEntry struct {
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

// DirResponse represents directory listing response
type DirResponse struct {
	Name     string     `json:"name"`
	Type     string     `json:"type"`
	Size     int64      `json:"size"`
	Path     string     `json:"path"`
	Contents []DirEntry `json:"contents"`
}

// Valid test users (demo only)
var validUsers = map[string]string{
	"admin": "admin123",
	"user":  "password",
}

var sessionMgr = &SessionManager{
	sessions: make(map[string]SessionData),
}

const (
	sessionDuration   = 1 * time.Hour
	sessionCookieName = "session"
	allowedBaseDir    = "" // Will be set to home directory at startup
)

// generateToken creates a cryptographically random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(username string) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sessions[token] = SessionData{
		Username:  username,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(sessionDuration),
	}

	return token, nil
}

// ValidateSession checks if a session token is valid and not expired
func (sm *SessionManager) ValidateSession(token string) (string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[token]
	if !exists {
		return "", fmt.Errorf("invalid session")
	}

	if time.Now().After(session.ExpiresAt) {
		return "", fmt.Errorf("session expired")
	}

	return session.Username, nil
}

// DestroySession removes a session
func (sm *SessionManager) DestroySession(token string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, token)
}

// CleanupExpiredSessions removes expired sessions
func (sm *SessionManager) CleanupExpiredSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for token, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, token)
		}
	}
}

// getSessionToken extracts session token from cookie
func getSessionToken(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// setSessionCookie sets a secure session cookie
func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// clearSessionCookie removes the session cookie
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// withJSON sets JSON response headers
func withJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

// securityHeaders adds security headers to response
func securityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
}

// handleLogin handles POST /api/login
func handleLogin(w http.ResponseWriter, r *http.Request) {
	securityHeaders(w)
	withJSON(w)

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Sanitize input
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Username and password required",
		})
		return
	}

	// Validate credentials
	expectedPassword, exists := validUsers[req.Username]
	if !exists || expectedPassword != req.Password {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Invalid username or password",
		})
		log.Printf("Failed login attempt for user: %s", html.EscapeString(req.Username))
		return
	}

	// Create session
	token, err := sessionMgr.CreateSession(req.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Failed to create session",
		})
		log.Printf("Error creating session: %v", err)
		return
	}

	setSessionCookie(w, token)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Message: "Login successful",
	})
	log.Printf("Successful login for user: %s", html.EscapeString(req.Username))
}

// handleLogout handles POST /api/logout
func handleLogout(w http.ResponseWriter, r *http.Request) {
	securityHeaders(w)
	withJSON(w)

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	token := getSessionToken(r)
	if token != "" {
		sessionMgr.DestroySession(token)
	}

	clearSessionCookie(w)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

// handleListDir handles GET /api/list
func handleListDir(w http.ResponseWriter, r *http.Request) {
	securityHeaders(w)
	withJSON(w)

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	// Validate session
	token := getSessionToken(r)
	username, err := sessionMgr.ValidateSession(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	// Get path from query parameter
	requestedPath := r.URL.Query().Get("path")
	if requestedPath == "" {
		requestedPath = "/"
	}

	// Validate path - prevent traversal attacks
	basePath, err := filepath.Abs(filepath.Join(allowedBaseDir, ".."))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Internal error",
		})
		return
	}

	cleanPath := filepath.Clean(filepath.Join(basePath, requestedPath))
	if !strings.HasPrefix(cleanPath, basePath) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Path traversal detected",
		})
		log.Printf("Path traversal attempt by %s: %s", username, requestedPath)
		return
	}

	// Read directory
	entries, err := os.ReadDir(cleanPath)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "Failed to read directory",
		})
		log.Printf("Error reading directory %s: %v", cleanPath, err)
		return
	}

	// Build response
	contents := make([]DirEntry, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // Skip entries we can't stat
		}

		entryType := "file"
		if entry.IsDir() {
			entryType = "dir"
		}

		contents = append(contents, DirEntry{
			Name:     entry.Name(),
			Type:     entryType,
			Size:     info.Size(),
			Modified: info.ModTime(),
		})
	}

	dirName := filepath.Base(cleanPath)
	if dirName == "." {
		dirName = "root"
	}

	response := DirResponse{
		Name:     dirName,
		Type:     "dir",
		Size:     0,
		Path:     cleanPath,
		Contents: contents,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	log.Printf("Directory access by %s: %s", username, cleanPath)
}

// handleRoot redirects root to /files or /login
func handleRoot(w http.ResponseWriter, r *http.Request) {
	token := getSessionToken(r)
	_, err := sessionMgr.ValidateSession(token)

	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/files", http.StatusSeeOther)
	}
}

// handleStatic serves static files from /static
func handleStatic(w http.ResponseWriter, r *http.Request) {
	securityHeaders(w)

	// Only serve from static directory
	path := r.URL.Path
	if strings.Contains(path, "..") {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Remove /static/ prefix
	filePath := strings.TrimPrefix(path, "/static/")
	filePath = filepath.Clean(filepath.Join("static", filePath))

	// Ensure file is within static directory
	staticDir, _ := filepath.Abs("./static")
	absPath, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absPath, staticDir) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Set correct MIME type based on file extension
	ext := filepath.Ext(filePath)
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}

	http.ServeFile(w, r, filePath)
}

// handlePages serves page files (login, files)
func handlePages(w http.ResponseWriter, r *http.Request) {
	securityHeaders(w)

	token := getSessionToken(r)
	_, sessionErr := sessionMgr.ValidateSession(token)

	path := r.URL.Path
	if path == "/login" {
		// Login page should be accessible to unauthenticated users
		http.ServeFile(w, r, "./static/login.html")
		return
	}

	if path == "/files" || strings.HasPrefix(path, "/files/") {
		// Files page requires authentication
		if sessionErr != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.ServeFile(w, r, "./static/files.html")
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

// filteringWriter filters out TLS handshake error messages
type filteringWriter struct {
	writer io.Writer
}

func (fw *filteringWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	if strings.Contains(msg, "TLS handshake error") && strings.Contains(msg, "unknown certificate") {
		return len(p), nil // Suppress this specific error
	}
	return fw.writer.Write(p)
}

func main() {
	// Cleanup expired sessions periodically
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			sessionMgr.CleanupExpiredSessions()
		}
	}()

	// Set allowed base directory to home
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	// Store home dir for path validation
	os.Setenv("ALLOWED_BASE_DIR", homeDir)

	// Register routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/logout", handleLogout)
	mux.HandleFunc("/api/list", handleListDir)
	mux.HandleFunc("/login", handlePages)
	mux.HandleFunc("/files", handlePages)
	mux.HandleFunc("/files/", handlePages)
	mux.HandleFunc("/static/", handleStatic)

	// TLS setup
	certFile := "server.crt"
	keyFile := "server.key"

	// Check if certificates exist, if not create them
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Printf("Certificates not found. Generate them with:")
		log.Printf("  openssl req -x509 -newkey rsa:4096 -keyout %s -out %s -days 365 -nodes -subj '/CN=localhost'", keyFile, certFile)
		log.Printf("Or run: make certs")
		return
	}

	log.Printf("Starting HTTPS server on https://localhost:8443")
	log.Printf("Test credentials:")
	log.Printf("  Username: admin, Password: admin123")
	log.Printf("  Username: user, Password: password")
	log.Printf("")
	log.Printf("NOTE: You will see 'TLS handshake error' messages - this is normal for self-signed certificates.")
	log.Printf("In your browser, click 'Advanced' and 'Proceed to localhost' to accept the certificate.")
	log.Printf("")

	server := &http.Server{
		Addr:     ":8443",
		Handler:  mux,
		ErrorLog: log.New(&filteringWriter{writer: os.Stderr}, "", log.LstdFlags),
	}

	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

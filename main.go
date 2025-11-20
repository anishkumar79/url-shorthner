package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type URLRecord struct {
	ID         int64     `json:"id"`
	ShortCode  string    `json:"short_code"`
	LongURL    string    `json:"long_url"`
	CreatedAt  time.Time `json:"created_at"`
	ClickCount int       `json:"click_count"`
}

type CreateURLRequest struct {
	URL string `json:"url"`
}

type CreateURLResponse struct {
	ShortURL  string `json:"short_url"`
	ShortCode string `json:"short_code"`
	LongURL   string `json:"long_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type App struct {
	DB *sql.DB
}

func NewApp() (*App, error) {
	db, err := sql.Open("sqlite3", "./urls.db")
	if err != nil {
		return nil, err
	}

	// Create urls table
	createUrlsTableSQL := `
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		short_code TEXT UNIQUE NOT NULL,
		long_url TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		click_count INTEGER DEFAULT 0
	);`

	_, err = db.Exec(createUrlsTableSQL)
	if err != nil {
		return nil, err
	}

	return &App{DB: db}, nil
}

func (app *App) generateShortCode() string {
	// Generate a random 6-character code
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 6

	b := make([]byte, codeLength)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to time-based generation if crypto/rand fails
		for i := 0; i < codeLength; i++ {
			b[i] = byte(time.Now().UnixNano() % int64(len(charset)))
			time.Sleep(time.Nanosecond)
		}
	}

	var code strings.Builder
	for i := 0; i < codeLength; i++ {
		code.WriteByte(charset[b[i]%byte(len(charset))])
	}
	return code.String()
}

func (app *App) createShortURL(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.URL == "" {
		respondWithError(w, http.StatusBadRequest, "URL is required")
		return
	}

	// Validate URL format
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		req.URL = "https://" + req.URL
	}

	// Generate unique short code
	var shortCode string
	for {
		shortCode = app.generateShortCode()
		var exists bool
		err := app.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = ?)", shortCode).Scan(&exists)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Database error")
			return
		}
		if !exists {
			break
		}
	}

	// Insert into database
	_, err := app.DB.Exec("INSERT INTO urls (short_code, long_url) VALUES (?, ?)", shortCode, req.URL)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create short URL")
		return
	}

	host := r.Host
	if host == "" {
		host = "localhost:8080"
	}
	shortURL := fmt.Sprintf("http://%s/%s", host, shortCode)

	response := CreateURLResponse{
		ShortURL:  shortURL,
		ShortCode: shortCode,
		LongURL:   req.URL,
	}

	w.Header().Set("Content-Type", "application/json")
	enableCORS(w)
	json.NewEncoder(w).Encode(response)
}

func (app *App) redirectShortURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["code"]

	var longURL string
	var clickCount int
	err := app.DB.QueryRow("SELECT long_url, click_count FROM urls WHERE short_code = ?", shortCode).Scan(&longURL, &clickCount)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Increment click count
	app.DB.Exec("UPDATE urls SET click_count = click_count + 1 WHERE short_code = ?", shortCode)

	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

func (app *App) getURLStats(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	vars := mux.Vars(r)
	shortCode := vars["code"]

	if shortCode == "" {
		respondWithError(w, http.StatusBadRequest, "Short code is required")
		return
	}

	var record URLRecord
	err := app.DB.QueryRow(
		"SELECT id, short_code, long_url, created_at, click_count FROM urls WHERE short_code = ?",
		shortCode,
	).Scan(&record.ID, &record.ShortCode, &record.LongURL, &record.CreatedAt, &record.ClickCount)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "URL not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enableCORS(w)
	json.NewEncoder(w).Encode(record)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (app *App) setupRoutes() *mux.Router {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/create", app.createShortURL).Methods("POST", "OPTIONS")
	api.HandleFunc("/stats/{code}", app.getURLStats).Methods("GET")

	// Serve static files
	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Serve root and common static files explicitly
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	r.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	r.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/style.css")
	})
	r.HandleFunc("/script.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/script.js")
	})

	// Redirect route (public) - only match 6-character alphanumeric codes
	r.HandleFunc("/{code:[a-zA-Z0-9]{6}}", app.redirectShortURL).Methods("GET")

	return r
}

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatal("Failed to initialize app:", err)
	}
	defer app.DB.Close()

	// Create static directory if it doesn't exist
	if _, err := os.Stat("static"); os.IsNotExist(err) {
		os.Mkdir("static", 0755)
	}

	port := os.Getenv("PORT")
    if port == "" {
     port = "8080" // fallback for local dev
}

log.Println("Server running on port:", port)
log.Fatal(http.ListenAndServe(":"+port, r))

}

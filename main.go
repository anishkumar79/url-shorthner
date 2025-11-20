package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	DB *sql.DB
}

type URL struct {
	ID        int    `json:"id"`
	ShortCode string `json:"short_code"`
	LongURL   string `json:"long_url"`
	Clicks    int    `json:"clicks"`
}

// --------------------- INIT APP -----------------------

func NewApp() (*App, error) {
	db, err := sql.Open("sqlite3", "./urls.db")
	if err != nil {
		return nil, err
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		short_code TEXT NOT NULL UNIQUE,
		long_url TEXT NOT NULL,
		clicks INTEGER DEFAULT 0
	);
	`
	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}

	return &App{DB: db}, nil
}

// --------------------- HELPERS ------------------------

func generateShortCode() string {
	rand.Seed(time.Now().UnixNano())
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	short := make([]byte, 6)
	for i := range short {
		short[i] = charset[rand.Intn(len(charset))]
	}
	return string(short)
}

// --------------------- HANDLERS -----------------------

func (app *App) HomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func (app *App) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	longURL := r.FormValue("long_url")

	if longURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	shortCode := generateShortCode()

	_, err := app.DB.Exec("INSERT INTO urls (short_code, long_url) VALUES (?, ?)", shortCode, longURL)
	if err != nil {
		http.Error(w, "Failed to store URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/success?short="+shortCode, http.StatusSeeOther)
}

func (app *App) RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortCode := mux.Vars(r)["code"]

	var longURL string
	err := app.DB.QueryRow("SELECT long_url FROM urls WHERE short_code = ?", shortCode).Scan(&longURL)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	app.DB.Exec("UPDATE urls SET clicks = clicks + 1 WHERE short_code = ?", shortCode)

	http.Redirect(w, r, longURL, http.StatusFound)
}

func (app *App) GetStats(w http.ResponseWriter, r *http.Request) {
	shortCode := mux.Vars(r)["code"]

	var url URL
	err := app.DB.QueryRow("SELECT id, short_code, long_url, clicks FROM urls WHERE short_code = ?", shortCode).
		Scan(&url.ID, &url.ShortCode, &url.LongURL, &url.Clicks)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(url)
}

// --------------------- ROUTES -------------------------

func (app *App) setupRoutes() *mux.Router {
	r := mux.NewRouter()

	// static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// UI
	r.HandleFunc("/", app.HomePage).Methods("GET")
	r.HandleFunc("/shorten", app.CreateShortURL).Methods("POST")

	// redirection
	r.HandleFunc("/{code}", app.RedirectURL).Methods("GET")

	// stats
	r.HandleFunc("/api/stats/{code}", app.GetStats).Methods("GET")

	return r
}

// --------------------- MAIN ---------------------------

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatal("Failed to initialize app:", err)
	}
	defer app.DB.Close()

	// Create static folder if missing
	if _, err := os.Stat("static"); os.IsNotExist(err) {
		os.Mkdir("static", 0755)
	}

	// IMPORTANT: setup routes
	r := app.setupRoutes()

	// Required for Render
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running on port:", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

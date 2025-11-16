# URL Shortener

A modern, full-featured URL shortener built with Go backend and a beautiful web interface.

## Features

- ✨ **Simple & Fast**: Shorten URLs instantly with a clean, modern interface
- 🔒 **Secure**: Uses cryptographically secure random code generation
- 📊 **Analytics**: Track click counts and creation dates for your shortened URLs
- 💾 **Persistent Storage**: SQLite database for reliable URL storage
- 📱 **Responsive Design**: Works perfectly on desktop and mobile devices
- 🎨 **Modern UI**: Beautiful gradient design with smooth animations
- 🌐 **No Authentication Required**: Public URL shortening - anyone can use it

## Prerequisites

- Go 1.21 or higher
- Git (optional)

## Installation

1. Clone or download this repository:
```bash
cd "url shortner"
```

2. Install dependencies:
```bash
go mod download
```

## Running the Application

1. Start the server:
```bash
go run main.go
```

2. Open your browser and navigate to:
```
http://localhost:8080
```

The server will automatically:
- Create the SQLite database (`urls.db`) if it doesn't exist
- Create the necessary database tables
- Serve the web interface and API

## Usage

1. **Shorten a URL**: Enter any long URL in the input field and click "Shorten"
2. **Copy Short URL**: Click the "Copy" button to copy the shortened URL to your clipboard
3. **View Statistics**: Click "View Stats" to see click counts and creation date
4. **Redirect**: Visit any shortened URL (e.g., `http://localhost:8080/abc123`) to be redirected to the original URL

## API Endpoints

### POST `/api/create`
Create a new short URL. **Public endpoint - no authentication required.**

**Request Body:**
```json
{
  "url": "https://example.com/very/long/url"
}
```

**Response:**
```json
{
  "short_url": "http://localhost:8080/abc123",
  "short_code": "abc123",
  "long_url": "https://example.com/very/long/url"
}
```

#### GET `/api/stats/{code}`
Get statistics for a shortened URL.

**Response:**
```json
{
  "id": 1,
  "short_code": "abc123",
  "long_url": "https://example.com/very/long/url",
  "created_at": "2025-11-16T13:00:00Z",
  "click_count": 42
}
```

#### GET `/{code}`
Redirect to the original URL associated with the short code.

## Configuration

You can set the port using the `PORT` environment variable:

```bash
# Windows PowerShell
$env:PORT="3000"; go run main.go

# Windows CMD
set PORT=3000 && go run main.go

# Linux/Mac
PORT=3000 go run main.go
```

Default port is `8080`.

## Project Structure

```
url shortner/
├── main.go              # Go backend server
├── go.mod              # Go module dependencies
├── go.sum              # Go module checksums (generated)
├── urls.db             # SQLite database (created on first run)
├── static/             # Frontend files
│   ├── index.html      # Main HTML page
│   ├── style.css       # Stylesheet
│   └── script.js       # Frontend JavaScript
└── README.md           # This file
```

## Technologies Used

- **Backend**: Go (Golang)
- **Database**: SQLite
- **Routing**: Gorilla Mux
- **Frontend**: HTML5, CSS3, JavaScript (Vanilla)
- **Storage**: SQLite database

## License

This project is open source and available for personal and commercial use.

## Contributing

Feel free to submit issues, fork the repository, and create pull requests for any improvements.


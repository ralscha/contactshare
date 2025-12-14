// Package main provides a web server for sharing contact information via vCard and QR codes.
package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/skip2/go-qrcode"
)

type Config struct {
	Name     string
	Bluesky  string
	Email    string
	Github   string
	Whatsapp string
	Facebook string
	Phone    string
	BaseURL  string
	Port     string
}

var (
	config       Config
	homeTemplate *template.Template
)

func loadConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	cfg := Config{
		Name:     os.Getenv("NAME"),
		Bluesky:  os.Getenv("BLUESKY"),
		Email:    os.Getenv("EMAIL"),
		Github:   os.Getenv("GITHUB"),
		Whatsapp: os.Getenv("WHATSAPP"),
		Facebook: os.Getenv("FACEBOOK"),
		Phone:    os.Getenv("PHONE"),
		BaseURL:  os.Getenv("BASE_URL"),
		Port:     os.Getenv("PORT"),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	// Validate required fields
	if cfg.Name == "" {
		return cfg, fmt.Errorf("NAME is required in .env")
	}

	if cfg.BaseURL == "" {
		return cfg, fmt.Errorf("BASE_URL is required in .env")
	}

	return cfg, nil
}

func initTemplates() error {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Name}} - Contact Info</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            max-width: 600px; 
            margin: 0 auto; 
            padding: 20px; 
            line-height: 1.6;
            background: #f5f5f5;
        }
        .container {
            background: white;
            border-radius: 8px;
            padding: 30px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 { 
            text-align: center;
            margin-bottom: 30px;
            color: #333;
            font-size: 2em;
        }
        .contact-list { 
            list-style: none; 
            padding: 0;
            margin-bottom: 20px;
        }
        .contact-list li { 
            margin-bottom: 15px;
            padding: 10px;
            background: #f9f9f9;
            border-radius: 5px;
            transition: background 0.2s;
        }
        .contact-list li:hover {
            background: #e9e9e9;
        }
        .contact-list strong {
            display: inline-block;
            min-width: 100px;
            color: #555;
        }
        .contact-list a { 
            text-decoration: none; 
            color: #007bff;
            word-break: break-all;
        }
        .contact-list a:hover {
            text-decoration: underline;
        }
        .btn { 
            display: block; 
            width: 100%; 
            padding: 15px; 
            background: #28a745; 
            color: white; 
            text-align: center; 
            text-decoration: none; 
            border-radius: 5px; 
            margin-top: 20px;
            font-weight: bold;
            transition: background 0.2s;
        }
        .btn:hover {
            background: #218838;
        }
        @media (max-width: 600px) {
            body { padding: 10px; }
            .container { padding: 20px; }
            h1 { font-size: 1.5em; }
            .contact-list strong {
                display: block;
                margin-bottom: 5px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{.Name}}</h1>
        <ul class="contact-list">
            {{if .Email}}<li><strong>Email:</strong> <a href="mailto:{{.Email}}">{{.Email}}</a></li>{{end}}
            {{if .Phone}}<li><strong>Phone:</strong> <a href="tel:{{.Phone}}">{{.Phone}}</a></li>{{end}}
            {{if .Bluesky}}<li><strong>Bluesky:</strong> <a href="{{.Bluesky}}" target="_blank" rel="noopener">Profile</a></li>{{end}}
            {{if .Github}}<li><strong>GitHub:</strong> <a href="{{.Github}}" target="_blank" rel="noopener">Profile</a></li>{{end}}
            {{if .Whatsapp}}<li><strong>WhatsApp:</strong> <a href="{{.Whatsapp}}" target="_blank" rel="noopener">Chat</a></li>{{end}}
            {{if .Facebook}}<li><strong>Facebook:</strong> <a href="{{.Facebook}}" target="_blank" rel="noopener">Profile</a></li>{{end}}
        </ul>
        <a href="/contact.vcf" class="btn">Download Contact (vCard)</a>
    </div>
</body>
</html>
`
	var err error
	homeTemplate, err = template.New("home").Parse(tmpl)
	return err
}

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next(w, r)
		log.Printf("Completed in %v", time.Since(start))
	}
}

func securityHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline'; img-src 'self' data:")
		next(w, r)
	}
}

func main() {
	var err error
	config, err = loadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	if err := initTemplates(); err != nil {
		log.Fatalf("Template initialization error: %v", err)
	}

	log.Printf("Starting server with contact info for: %s", config.Name)

	http.HandleFunc("/", securityHeaders(loggingMiddleware(homeHandler)))
	http.HandleFunc("/qr", loggingMiddleware(qrHandler))
	http.HandleFunc("/contact.vcf", loggingMiddleware(vcardHandler))
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)

	srv := &http.Server{
		Addr:         ":" + config.Port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"status":"healthy","service":"contactshare"}`)
}

func faviconHandler(w http.ResponseWriter, _ *http.Request) {
	// Simple 1x1 transparent PNG
	favicon := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
		0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=604800")
	_, _ = w.Write(favicon)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if err := homeTemplate.Execute(w, config); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func qrHandler(w http.ResponseWriter, _ *http.Request) {
	png, err := qrcode.Encode(config.BaseURL, qrcode.Medium, 256)
	if err != nil {
		log.Printf("QR code generation error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write(png)
}

func vcardHandler(w http.ResponseWriter, _ *http.Request) {
	var vcard strings.Builder
	vcard.WriteString("BEGIN:VCARD\nVERSION:3.0\n")
	vcard.WriteString(fmt.Sprintf("FN:%s\n", config.Name))

	if config.Email != "" {
		vcard.WriteString(fmt.Sprintf("EMAIL:%s\n", config.Email))
	}
	if config.Phone != "" {
		vcard.WriteString(fmt.Sprintf("TEL;TYPE=CELL:%s\n", config.Phone))
	}
	if config.BaseURL != "" {
		vcard.WriteString(fmt.Sprintf("URL:%s\n", config.BaseURL))
	}
	if config.Github != "" {
		vcard.WriteString(fmt.Sprintf("URL:%s\n", config.Github))
	}
	if config.Facebook != "" {
		vcard.WriteString(fmt.Sprintf("URL:%s\n", config.Facebook))
	}
	if config.Bluesky != "" {
		vcard.WriteString(fmt.Sprintf("URL:%s\n", config.Bluesky))
	}
	if config.Whatsapp != "" {
		vcard.WriteString(fmt.Sprintf("URL:%s\n", config.Whatsapp))
	}

	// Add notes for context
	var notes strings.Builder
	if config.Bluesky != "" {
		notes.WriteString("Bluesky: " + config.Bluesky + "\\n")
	}
	if config.Github != "" {
		notes.WriteString("GitHub: " + config.Github + "\\n")
	}
	if config.Facebook != "" {
		notes.WriteString("Facebook: " + config.Facebook + "\\n")
	}
	if notes.Len() > 0 {
		vcard.WriteString(fmt.Sprintf("NOTE:%s\n", notes.String()))
	}

	vcard.WriteString("END:VCARD")

	w.Header().Set("Content-Type", "text/vcard")
	w.Header().Set("Content-Disposition", "attachment; filename=\"contact.vcf\"")
	_, _ = w.Write([]byte(vcard.String()))
}

package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
)

var (
	//go:embed all:templates/*
	templateFS embed.FS

	//go:embed htmx.min.js
	htmx embed.FS

	// Parsed templates
	html *template.Template
)

func main() {
	var err error
	htmlFS, _ := fs.Sub(templateFS, "templates")
	html, err = template.ParseFS(htmlFS, "*.html")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", http.HandlerFunc(index))
	mux.Handle("GET /htmx.min.js", http.FileServer(http.FS(htmx)))
	mux.Handle("GET /details/{id}", http.HandlerFunc(updateDetails))
	log.Fatal(http.ListenAndServe(":8080", handlers.RecoveryHandler()(mux)))
}

func index(w http.ResponseWriter, r *http.Request) {
	table := []int{1, 2, 3, 4}
	arg := map[string]any{"table": table, "details": table[0]}
	if err := html.ExecuteTemplate(w, "index.html", arg); err != nil {
		panic(err)
	}
}

func updateDetails(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := html.ExecuteTemplate(w, "details.html", id); err != nil {
		panic(err)
	}
}

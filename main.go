package main

import (
	"embed"
	"fmt"
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

	//go:embed bootstrap.min.css
	bootstrap embed.FS

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
	mux.Handle("GET /bootstrap.min.css", http.FileServer(http.FS(bootstrap)))
	mux.Handle("GET /details/{id}", http.HandlerFunc(updateDetails))
	log.Fatal(http.ListenAndServe(":8080", handlers.RecoveryHandler()(mux)))
}

func index(w http.ResponseWriter, r *http.Request) {
	table := []string{}
	for i := range 19999 {
		table = append(table, fmt.Sprintf("Service-%d", i+1))
	}
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

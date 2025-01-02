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
	css embed.FS

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
	mux.Handle("GET /bootstrap.min.css", http.FileServer(http.FS(css)))
	mux.Handle("GET /details/{service}", http.HandlerFunc(updateDetails))
	mux.Handle("GET /showUsers/{state}", http.HandlerFunc(showUsers))
	log.Fatal(http.ListenAndServe(":8080", handlers.RecoveryHandler()(mux)))
}

func index(w http.ResponseWriter, r *http.Request) {
	table := []string{}
	for i := range 20 {
		table = append(table, fmt.Sprintf("Service-%d", i+1))
	}
	arg := map[string]any{
		"table": table,
		"details": &serviceDetails{
			Service:   table[0],
			ShowUsers: "off",
		},
	}
	if err := html.ExecuteTemplate(w, "index.html", arg); err != nil {
		panic(err)
	}
}

type serviceDetails struct {
	Service   string
	ShowUsers string
}

func updateDetails(w http.ResponseWriter, r *http.Request) {
	details := &serviceDetails{
		Service:   r.PathValue("service"),
		ShowUsers: r.URL.Query().Get("showUsers"),
	}
	if err := html.ExecuteTemplate(w, "details.html", details); err != nil {
		panic(err)
	}
	fmt.Fprintf(w,
		`<input type="hidden" id="%s" name="%s" hx-swap-oob="true" value="%s" />`,
		"service", "service", details.Service)
}

func showUsers(w http.ResponseWriter, r *http.Request) {
	details := &serviceDetails{
		Service:   r.URL.Query().Get("service"),
		ShowUsers: r.PathValue("state"),
	}
	if err := html.ExecuteTemplate(w, "details.html", details); err != nil {
		panic(err)
	}
	fmt.Fprintf(w,
		`<input type="hidden" id="%s" name="%s" hx-swap-oob="true" value="%s" />`,
		"showUsers", "showUsers", details.ShowUsers)
}

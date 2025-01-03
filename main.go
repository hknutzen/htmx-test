package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"

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
	mux.Handle("GET /htmx.min.js", http.FileServer(http.FS(htmx)))
	mux.Handle("GET /bootstrap.min.css", http.FileServer(http.FS(css)))
	mux.Handle("GET /", http.HandlerFunc(index))
	mux.Handle("GET /serviceList/{type}", http.HandlerFunc(updateServiceList))
	mux.Handle("GET /details/{service}", http.HandlerFunc(updateDetails))
	mux.Handle("GET /showUsers/{state}", http.HandlerFunc(showUsers))
	log.Fatal(http.ListenAndServe(":8080", handlers.RecoveryHandler()(mux)))
}

type serviceDetails struct {
	Service   string
	ShowUsers string
}

func index(w http.ResponseWriter, r *http.Request) {
	table := getServiceList("owner")
	first := ""
	if len(table) > 0 {
		first = table[0]
	}
	args := map[string]any{
		"table": table,
		"details": &serviceDetails{
			Service:   first,
			ShowUsers: "off",
		},
	}
	if err := html.ExecuteTemplate(w, "index.html", args); err != nil {
		panic(err)
	}
}

func updateServiceList(w http.ResponseWriter, r *http.Request) {
	listType := r.PathValue("type")
	table := getServiceList(listType)
	if err := html.ExecuteTemplate(w, "table.html", table); err != nil {
		panic(err)
	}
	if len(table) == 0 {
		fmt.Fprintln(w, `<div id="details" hx-swap-oob="true"></div>`)
	} else {
		details := &serviceDetails{
			Service:   table[0],
			ShowUsers: r.URL.Query().Get("showUsers"),
		}
		fmt.Fprintln(w, `<div id="details" hx-swap-oob="true">`)
		if err := html.ExecuteTemplate(w, "details.html", details); err != nil {
			panic(err)
		}
		fmt.Fprintln(w, `</div`)
		fmt.Fprintf(w,
			`<input type="hidden" id="%s" name="%s" hx-swap-oob="true" value="%s" />`,
			"service", "service", details.Service)
	}
}

func getServiceList(listType string) []string {
	size := 0
	switch listType {
	case "owner":
		size = 20
	case "user":
		size = 10000
	}
	prefix := strings.ToUpper(listType)
	table := []string{}
	for i := range size {
		table = append(table, fmt.Sprintf("%s-Service-%d", prefix, i+1))
	}
	return table
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

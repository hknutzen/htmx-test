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

type templateParams struct {
	Table   []string
	Details *serviceDetails
}
type serviceDetails struct {
	Name        string
	Description string
	Owner       string
	ShowUsers   string
	Users       []*user
}
type user struct {
	Name  string
	IP    string
	Owner string
}

func index(w http.ResponseWriter, r *http.Request) {
	params := getServiceListParams("owner")
	if err := html.ExecuteTemplate(w, "index.html", params); err != nil {
		panic(err)
	}
}

func updateServiceList(w http.ResponseWriter, r *http.Request) {
	listType := r.PathValue("type")
	params := getServiceListParams(listType)
	if err := html.ExecuteTemplate(w, "table.html", params); err != nil {
		panic(err)
	}
	if len(params.Table) == 0 {
		fmt.Fprintln(w, `<div id="details" hx-swap-oob="true"></div>`)
	} else {
		details := params.Details
		details.ShowUsers = r.URL.Query().Get("showUsers")
		fmt.Fprintln(w, `<div id="details" hx-swap-oob="true">`)
		if err := html.ExecuteTemplate(w, "details.html", details); err != nil {
			panic(err)
		}
		fmt.Fprintln(w, `</div`)
	}
}

func getServiceListParams(listType string) templateParams {
	size := 0
	switch listType {
	case "owner":
		size = 20
	case "user":
		size = 21845
	}
	prefix := strings.ToUpper(listType)
	table := []string{}
	for i := range size {
		table = append(table, fmt.Sprintf("%s-Service-%d", prefix, i+1))
	}
	var details serviceDetails
	if len(table) > 0 {
		details = getDetails(table[0])
	}
	details.ShowUsers = "off"
	return templateParams{
		Table:   table,
		Details: &details,
	}
}

func getDetails(name string) serviceDetails {
	if name == "" {
		return serviceDetails{}
	}
	users := make([]*user, 7)
	for i := range users {
		u := &user{
			Name:  fmt.Sprintf("host:h%d-of-%s", i+1, name),
			IP:    fmt.Sprintf("10.1.2.%d", i+1),
			Owner: fmt.Sprintf("Owner-%d", i+1),
		}
		users[i] = u
	}
	return serviceDetails{
		Name:        name,
		Description: "Description of " + name,
		Owner:       "Owner-" + name,
		Users:       users,
	}
}

func updateDetails(w http.ResponseWriter, r *http.Request) {
	details := getDetails(r.PathValue("service"))
	details.ShowUsers = r.URL.Query().Get("showUsers")
	if err := html.ExecuteTemplate(w, "details.html", details); err != nil {
		panic(err)
	}
	fmt.Fprintf(w,
		`<input type="hidden" id="%s" name="%s" hx-swap-oob="true" value="%s" />`,
		"service", "service", details.Name)
}

func showUsers(w http.ResponseWriter, r *http.Request) {
	details := getDetails(r.URL.Query().Get("service"))
	details.ShowUsers = r.PathValue("state")
	if err := html.ExecuteTemplate(w, "details.html", details); err != nil {
		panic(err)
	}
	fmt.Fprintf(w,
		`<input type="hidden" id="%s" name="%s" hx-swap-oob="true" value="%s" />`,
		"showUsers", "showUsers", details.ShowUsers)
}

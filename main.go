package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/handlers"
)

var (
	//go:embed all:templates/*
	templateFS embed.FS
	//go:embed htmx.min.js
	htmx embed.FS
	//go:embed bootstrap.bundle.min.js
	bootstrapJS embed.FS
	//go:embed bootstrap.min.css
	bootstrapCSS embed.FS
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
	mux.Handle("GET /bootstrap.bundle.min.js",
		http.FileServer(http.FS(bootstrapJS)))
	mux.Handle("GET /bootstrap.min.css", http.FileServer(http.FS(bootstrapCSS)))

	mux.Handle("GET /", http.HandlerFunc(index))
	mux.Handle("GET /services/{type}", http.HandlerFunc(updateServices))
	mux.Handle("GET /details/{service}", http.HandlerFunc(updateDetails))
	mux.Handle("GET /showUsers/{state}", http.HandlerFunc(showUsers))
	mux.Handle("GET /admins/{owner}", http.HandlerFunc(updateAdmins))
	log.Fatal(http.ListenAndServe(":8080", handlers.RecoveryHandler()(mux)))
}

type templateParams struct {
	ServiceType string
	ServiceList []string
	Details     *serviceDetails
}
type serviceDetails struct {
	Name        string
	Description string
	Owner       string
	QueryParams string
	ShowUsers   string
	Users       []*user
	UOwner      string
	Admins      []string
}
type user struct {
	Name  string
	IP    string
	Owner string
}

func index(w http.ResponseWriter, r *http.Request) {
	params := getServiceListParams("user")
	if err := html.ExecuteTemplate(w, "index.html", params); err != nil {
		panic(err)
	}
}

func updateServices(w http.ResponseWriter, r *http.Request) {
	serviceType := r.PathValue("type")
	params := getServiceListParams(serviceType)
	params.Details.ShowUsers = r.URL.Query().Get("showUsers")
	params.Details.QueryParams, _ = url.QueryUnescape(r.URL.String())
	if err := html.ExecuteTemplate(w, "services.html", params); err != nil {
		panic(err)
	}
}

func updateDetails(w http.ResponseWriter, r *http.Request) {
	dt := getDetails(r.PathValue("service"))
	dt.ShowUsers = r.URL.Query().Get("showUsers")
	if err := html.ExecuteTemplate(w, "service-details.html", dt); err != nil {
		panic(err)
	}
	setHiddenOOB(w, "service", dt.Name)
}

func showUsers(w http.ResponseWriter, r *http.Request) {
	dt := getDetails(r.URL.Query().Get("service"))
	dt.ShowUsers = r.PathValue("state")
	if err := html.ExecuteTemplate(w, "service-details.html", dt); err != nil {
		panic(err)
	}
}

func updateAdmins(w http.ResponseWriter, r *http.Request) {
	uOwner := r.PathValue("owner")
	admins := getAdmins(uOwner)
	details := &serviceDetails{
		UOwner: uOwner,
		Admins: admins,
	}
	if err := html.ExecuteTemplate(w, "admins.html", details); err != nil {
		panic(err)
	}
}

func setHiddenOOB(w http.ResponseWriter, name, value string) {
	fmt.Fprintf(w,
		`<input type="hidden" id="%s" name="%s" hx-swap-oob="true" value="%s" />`,
		name, name, value)
}

func getServiceListParams(serviceType string) templateParams {
	size := 0
	var prefix string
	switch serviceType {
	case "owner":
		prefix = "Genutzter"
		size = 21845
	case "user":
		prefix = "Eigener"
		size = 20
	case "visible":
		prefix = "Nutzbarer"
	case "search":
		prefix = "Gesuchter"
		size = 5
	}
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
		ServiceType: serviceType,
		ServiceList: table,
		Details:     &details,
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
	uOwner := users[0].Owner
	admins := getAdmins(uOwner)
	return serviceDetails{
		Name:        name,
		Description: "Description of " + name,
		Owner:       "Owner-" + name,
		Users:       users,
		UOwner:      uOwner,
		Admins:      admins,
	}
}

func getAdmins(owner string) []string {
	if owner == "" {
		return nil
	}
	c := owner[len(owner)-1:]
	i, _ := strconv.Atoi(c)
	if i == 0 {
		i = 1
	}
	admins := make([]string, i)
	for i := range admins {
		admins[i] = fmt.Sprintf("admin-%d@example.com", i+1)
	}
	return admins
}

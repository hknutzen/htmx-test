package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"

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
	params := getServiceListParams("owner")
	if err := html.ExecuteTemplate(w, "index.html", params); err != nil {
		panic(err)
	}
}

func updateServiceList(w http.ResponseWriter, r *http.Request) {
	serviceType := r.PathValue("type")
	params := getServiceListParams(serviceType)
	if err := html.ExecuteTemplate(w, "service-list.html", params); err != nil {
		panic(err)
	}
	details := params.Details
	details.ShowUsers = r.URL.Query().Get("showUsers")
	fmt.Fprintln(w, `<div id="details" hx-swap-oob="innerHTML ">`)
	if err := html.ExecuteTemplate(w, "details.html", details); err != nil {
		panic(err)
	}
	fmt.Fprintln(w, `</div`)
}

func getServiceListParams(serviceType string) templateParams {
	size := 0
	var prefix string
	switch serviceType {
	case "owner":
		prefix = "Eigener"
		size = 20
	case "user":
		prefix = "Genutzter"
		size = 21845
	case "visible":
		prefix = "Nutzbarer"
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

func updateDetails(w http.ResponseWriter, r *http.Request) {
	details := getDetails(r.PathValue("service"))
	details.ShowUsers = r.URL.Query().Get("showUsers")
	if err := html.ExecuteTemplate(w, "details.html", details); err != nil {
		panic(err)
	}
	setHiddenOOB(w, "service", details.Name)
}

func showUsers(w http.ResponseWriter, r *http.Request) {
	details := getDetails(r.URL.Query().Get("service"))
	details.ShowUsers = r.PathValue("state")
	if err := html.ExecuteTemplate(w, "details.html", details); err != nil {
		panic(err)
	}
	setHiddenOOB(w, "showUsers", details.ShowUsers)
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

func setHiddenOOB(w http.ResponseWriter, name, value string) {
	fmt.Fprintf(w,
		`<input type="hidden" id="%s" name="%s" hx-swap-oob="true" value="%s" />`,
		name, name, value)
}

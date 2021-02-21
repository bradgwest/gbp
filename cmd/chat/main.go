package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/bradgwest/gbp/pkg/auth"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
)

// templateHandler represents a simple template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP Request
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	t.templ.Execute(w, r)
}

func main() {
	var addr = flag.String("addr", ":8080", "The address of the application")
	flag.Parse()

	goth.UseProviders(
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), fmt.Sprintf("http://localhost%s/auth/callback/github", *addr)),
	)

	r := newRoom()
	http.Handle("/chat", auth.Must(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", auth.LoginHandler)
	http.Handle("/room", r)

	// start the room
	go r.run()

	// start the webserver
	log.Println("Starting the webserver on:", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

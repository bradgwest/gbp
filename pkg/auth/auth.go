package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/markbates/goth/gothic"
	"github.com/stretchr/objx"
)

type authHandler struct {
	next http.Handler
}

// ServeHTTP handles authentication
func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("auth")

	if err == http.ErrNoCookie {
		// not authenticated
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.next.ServeHTTP(w, r)
}

// Must returns an authentication handler with the given next handler
func Must(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

// LoginHandler handles the third-party login process
// format: /auth/{action}/{provider}
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]

	switch action {
	case "login":
		// Add the provider to the query string. This is a hack.
		q := r.URL.Query()
		q.Add("provider", provider)
		r.URL.RawQuery = q.Encode()
		gothic.BeginAuthHandler(w, r)
	case "callback":
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to get user from %s: %s", provider, err), http.StatusInternalServerError)
			return
		}

		authCookieValue := objx.New(map[string]string{"name": user.Name}).MustBase64()
		http.SetCookie(w, &http.Cookie{Name: "auth", Value: authCookieValue, Path: "/"})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supported", action)
	}
}

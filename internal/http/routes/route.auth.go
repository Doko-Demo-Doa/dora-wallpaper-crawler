package routes

import (
	"crypto/subtle"
	"net/http"
)

// requireToken guards a route group with a static bearer token.
// An empty token disables the group entirely (fail closed).
func requireToken(token string) func(http.Handler) http.Handler {
	want := []byte("Bearer " + token)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := []byte(r.Header.Get("Authorization"))
			if token == "" || subtle.ConstantTimeCompare(got, want) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

/*
Package middleware defines fuctions to be run on entry to the app, before the requested
endpoint is called.
It will generally either:
	a) move the end user onto the requested endpoint
	b) stop execution, usually with an error http.Statuscode.
*/
package middleware

import (
	"goafweb"
	"goafweb/context"
	"net/http"
	"strings"
)

type JsonAuthMW struct {
	UserService goafweb.UserService
}

// CheckUser will check the users Authorization header and then check to see if a user exists
// in the database.  If it does, the User is added to the request Context.
func (mw *JsonAuthMW) CheckUser(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")
		if len(bearer) < len("Bearer") {
			next.ServeHTTP(w, r)
			return
		}
		token := strings.TrimSpace(bearer[len("Bearer"):])
		user, err := mw.UserService.GetByRemember(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		r = r.WithContext(context.WithUser(r.Context(), user))
		next.ServeHTTP(w, r)
	}
}

// RequireUser will check that a user is set in the request context.
// It if is, the requested handler will be called.
// If not,  the server responsds with a http.StatusUnauthorized header and further execution is stopped.
func (mw *JsonAuthMW) RequireUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := context.GetUser(r.Context())
		if user == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

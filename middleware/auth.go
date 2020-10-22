package middleware

import (
	"goafweb"
	"goafweb/context"
	"net/http"
)

type authMW struct {
	UserService goafweb.UserService
}

func NewAuthMW(us goafweb.UserService) *authMW {
	return &authMW{
		UserService: us,
	}
}

// CheckUser will check the users http Cookies and then check to see if a user exists
// in the database.  If it does, the User is added to the request Context.
func (mw *authMW) CheckUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rememberCookie, err := r.Cookie("rememberToken")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		user, err := mw.UserService.GetByRemember(rememberCookie.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		r = r.WithContext(context.WithUser(r.Context(), user))

		next.ServeHTTP(w, r)
	})
}

// RequireUser will check that a user is set in the request context.
// It if is, the requested handler will be called.
// If not,  the server responsds with a redirect to the login page and further execution is stopped.
func (mw *authMW) RequireUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := context.GetUser(r.Context())

		if user == nil {
			http.Redirect(w, r, "/login", http.StatusUnauthorized)
		}
		next(w, r)
	}
}

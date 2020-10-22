package handlers

import (
	"goafweb/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

type app struct {
	authMW   middleware.AuthMW
	users    *userHandler
	articles *articleHandler
	router   *mux.Router
}

func NewApp(auth middleware.AuthMW, uh *userHandler, ah *articleHandler, r *mux.Router) *app {
	app := &app{
		authMW:   auth,
		users:    uh,
		articles: ah,
		router:   r,
	}
	app.routes()
	return app
}

func (a *app) routes() {
	r := a.router
	r.Use(a.authMW.CheckUser)
	// /api/user
	r.HandleFunc("/user", a.authMW.RequireUser(a.users.Create)).Methods(http.MethodPost)
	r.HandleFunc("/signup",a.users.Create).Methods("POST")
	r.HandleFunc("/login", a.users.Login).Methods("POST")
	r.HandleFunc("/logout", a.users.Logout).Methods("GET")
	r.HandleFunc("/forgot", a.users.Forgot).Methods("POST")
	r.HandleFunc("/reset", a.users.Reset).Methods("POST")

	// /api/article/
	r.HandleFunc("/article/{id:[0-9]+}", a.articles.View).Methods(http.MethodGet)
	r.HandleFunc("/article", a.articles.Create).Methods(http.MethodPost)
	r.HandleFunc("/article", a.authMW.RequireUser(a.articles.Update)).Methods(http.MethodPut)
	r.HandleFunc("/article", a.articles.Delete).Methods(http.MethodDelete)
}

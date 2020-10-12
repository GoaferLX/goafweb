/*
Package main pieces together all the components of the app and runs the binary.
*/
package main

import (
	"flag"
	"fmt"
	"goafweb/handlers"
	"goafweb/middleware"
	"goafweb/rand"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

func main() {
	prod := flag.Bool("prod", false, "Provide this flag in production. This ensures that a .config file is provided before the application starts.")
	flag.Parse()

	cfg := LoadConfig(*prod)
	dbcfg := cfg.Database
	mgcfg := cfg.Mailgun

	services, err := NewServices(
		WithGorm(dbcfg.Dialect, dbcfg.dsn(), cfg.isProd()),
		WithUsers(cfg.PWPepper, cfg.HMACKey),
		WithArticles(),
		WithMail(mgcfg.Domain, mgcfg.APIKey),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := services.AutoMigrate(); err != nil {
		log.Fatal("Could not initiate database tables")
	}

	users := handlers.NewUsers(services.UserService, services.MailService)
	articles := handlers.NewArticles(services.ArticleService)

	// Inititate middlewares

	authMW := middleware.JsonAuthMW{UserService: services.UserService}
	//	_ = authMW

	csrfbytes, err := rand.Bytes(32)
	if err != nil {
		log.Fatal(err)
	}
	csrfmw := csrf.Protect(csrfbytes, csrf.Secure(cfg.isProd()))
	_ = csrfmw

	r := mux.NewRouter().StrictSlash(true)
	api := r.PathPrefix("/api/").Subrouter()
	// User routes
	api.HandleFunc("/user", authMW.RequireUser(users.Create)).Methods(http.MethodPost)
	api.HandleFunc("/login", users.Login).Methods("POST")
	api.HandleFunc("/logout", users.Logout).Methods("GET")

	api.HandleFunc("/forgot", users.Forgot).Methods("POST")
	api.HandleFunc("/reset", users.Reset).Methods("POST")

	api.HandleFunc("/article/{id:[0-9]+}", articles.View).Methods(http.MethodGet)
	api.HandleFunc("/article", articles.Create).Methods(http.MethodPost)
	api.HandleFunc("/article", authMW.RequireUser(articles.Update)).Methods(http.MethodPut)
	api.HandleFunc("/article", articles.Delete).Methods(http.MethodDelete)

	//server := &http.Server{Handler: api, Addr: fmt.Sprintf(":%d", cfg.Port)}
	//log.Fatal(server.ListenAndServe())
	log.Printf("Server listening on port: %d", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), authMW.CheckUser(r)))

}

/*
Package main pieces together all the components of the app and runs the binary.
*/
package main

import (
	"flag"
	"fmt"
	"goafweb/handlers"
	"goafweb/middleware"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
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
		log.Fatalf("Could not initiate app: %s", err)
	}
	if err := services.AutoMigrate(); err != nil {
		log.Fatalf("Could not initiate database tables: %s", err)
	}

	router := mux.NewRouter().StrictSlash(true)

	handlers.NewApp(
		middleware.NewJsonAuthMW(services.UserService),
		handlers.NewUsers(services.UserService, services.MailService),
		handlers.NewArticles(services.ArticleService),
		router,
	)
	server := &http.Server{Handler: router, Addr: fmt.Sprintf(":%d", cfg.Port)}
	log.Printf("Server listening on port: %d", cfg.Port)
	log.Fatal(server.ListenAndServe())
	//log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), authMW.CheckUser(r)))

}

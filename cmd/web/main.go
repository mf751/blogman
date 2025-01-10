package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v4/stdlib"
	"golang.org/x/crypto/bcrypt"

	"github.com/mf751/blogman/interanl/models"
)

type application struct {
	errLogger     *log.Logger
	infoLogger    *log.Logger
	users         models.UsersModel
	blogs         models.BlogsModel
	templateCache map[string]*template.Template
}

func main() {
	infoLogger := log.New(os.Stdout, "[info ]\t", log.Ldate|log.Ltime)
	errLogger := log.New(os.Stdout, "[info ]\t", log.Ldate|log.Ltime)

	address := flag.String("addr", ":4001", "HTTP network address")
	dsn := flag.String(
		"dsn",
		"postgres://postgres:1319@localhost:5432/blogman",
		"Postgresql data source name",
	)
	flag.Parse()

	db, err := openDB(*dsn)
	if err != nil {
		errLogger.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errLogger.Fatal(err)
	}

	app := application{
		errLogger:     errLogger,
		infoLogger:    infoLogger,
		users:         models.UsersModel{DB: db},
		blogs:         models.BlogsModel{DB: db},
		templateCache: templateCache,
	}

	user := models.User{}
	user.Name = "John Doe"
	user.Email = "johndose@gmail.com"
	user.UserName = "johndoe"
	user.HashedPassword, err = bcrypt.GenerateFromPassword([]byte("johndoe"), 12)
	_, err = app.users.Insert(user)
	if err != nil {
		log.Fatal(err)
	}
	user.Name = "Lane wagnar"
	user.Email = "lanewagnar@gmail.com"
	user.UserName = "lanewagnar"
	user.HashedPassword, err = bcrypt.GenerateFromPassword([]byte("lanewagnar"), 12)
	_, err = app.users.Insert(user)
	if err != nil {
		log.Fatal(err)
	}
	user.Name = "moshrif"
	user.Email = "moshrif@gmail.com"
	user.UserName = "moshrif"
	user.HashedPassword, err = bcrypt.GenerateFromPassword([]byte("moshrif"), 12)
	_, err = app.users.Insert(user)
	if err != nil {
		log.Fatal(err)
	}
	user.Name = "falah"
	user.Email = "falah@gmail.com"
	user.UserName = "falah"
	user.HashedPassword, err = bcrypt.GenerateFromPassword([]byte("falah"), 12)
	_, err = app.users.Insert(user)
	if err != nil {
		log.Fatal(err)
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	server := &http.Server{
		Addr:         *address,
		Handler:      app.mainMux(),
		ErrorLog:     errLogger,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig:    tlsConfig,
	}

	infoLogger.Printf("Starting server on %v", *address)
	err = server.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	if err != nil {
		errLogger.Fatal(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	sql.Register("postgres", stdlib.GetDefaultDriver())
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

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

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/jackc/pgx/v4/stdlib"

	"github.com/mf751/blogman/interanl/models"
)

type application struct {
	errLogger      *log.Logger
	infoLogger     *log.Logger
	users          models.UsersModel
	blogs          models.BlogsModel
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
	formDecoder    *form.Decoder
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

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	app := application{
		errLogger:      errLogger,
		infoLogger:     infoLogger,
		users:          models.UsersModel{DB: db},
		blogs:          models.BlogsModel{DB: db},
		templateCache:  templateCache,
		sessionManager: sessionManager,
		formDecoder:    formDecoder,
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

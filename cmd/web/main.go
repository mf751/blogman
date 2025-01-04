package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v4/stdlib"
)

type application struct {
	errLogger  *log.Logger
	infoLogger *log.Logger
	db         *sql.DB
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

	app := application{
		errLogger:  errLogger,
		infoLogger: infoLogger,
		db:         db,
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

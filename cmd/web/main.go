package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"time"
)

type application struct {
	errLogger  *log.Logger
	infoLogger *log.Logger
}

func main() {
	infoLogger := log.New(os.Stdout, "[info ]\t", log.Ldate|log.Ltime)
	errLogger := log.New(os.Stdout, "[info ]\t", log.Ldate|log.Ltime)
	app := application{
		errLogger:  errLogger,
		infoLogger: infoLogger,
	}

	address := flag.String("addr", ":4001", "HTTP network address")
	flag.Parse()

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
	err := server.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	if err != nil {
		errLogger.Fatal(err)
	}
}

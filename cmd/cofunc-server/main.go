package main

import (
	"crypto/tls"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	if err := serve(); err != nil {
		panic(err)
	}
}

// Serve starts a server on the given address
func serve() error {
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           InitMux(),
		TLSConfig:         &tls.Config{},
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
	}
	return srv.ListenAndServe()
}

// InitMux initializes the mux for the server
func InitMux() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/v1/flow/", &FlowHandler{})
	r.Handle("/v1/flows/", &FlowsHandler{})
	return r
}

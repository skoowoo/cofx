package main

import (
	"crypto/tls"
	"net/http"

	"github.com/cofunclabs/cofunc/service"
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
	svc := service.New()
	r := mux.NewRouter()
	r.Handle("/v1/flows/list", &FlowListHandler{svc}).Methods("GET").Schemes("http")
	r.Handle("/v1/flows/run", &FlowRunHandler{svc}).Methods("PUT").Schemes("http")
	r.Handle("/v1/flows/add", &FlowAddHandler{svc}).Methods("POST").Schemes("http")
	r.Handle("/v1/flows/{id}/ready", &FlowReadyHandler{svc}).Methods("PUT").Schemes("http")
	r.Handle("/v1/flows/{id}/start", &FlowStartHandler{svc}).Methods("PUT").Schemes("http")
	r.Handle("/v1/flows/{id}/stop", &FlowStopHandler{svc}).Methods("PUT").Schemes("http")
	r.Handle("/v1/flows/{id}/delete", nil).Methods("PUT").Schemes("http")
	r.Handle("/v1/flows/{id}/status", &FlowStatusHandler{svc}).Methods("GET").Schemes("http")
	return r
}

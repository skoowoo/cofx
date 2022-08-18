package main

import "net/http"

type FlowsHandler struct {
}

func (h *FlowsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// /v1/flows/list
}

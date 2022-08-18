package main

import "net/http"

// FlowHandler is a http.Handler for the flow service
type FlowHandler struct {
}

func (h *FlowHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// /v1/flow/add
	// /v1/flow/<flowid>/ready
	// /v1/flow/<flowid>/start
	// /v1/flow/<flowid>/stop
	// /v1/flow/<flowid>/delete
	// /v1/flow/<flowid>/inspect
}

func (h *FlowHandler) AddFlow() error {
	return nil
}

func (h *FlowHandler) ReadyFlow() error {
	return nil
}

func (h *FlowHandler) StartFlow() error {
	return nil
}

func (h *FlowHandler) StopFlow() error {
	return nil
}

func (h *FlowHandler) DeleteFlow() error {
	return nil
}

func (h *FlowHandler) InspectFlow() error {
	return nil
}

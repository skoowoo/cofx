package main

import (
	"context"
	"log"
	"net/http"

	"github.com/cofunclabs/cofunc/pkg/nameid"
	"github.com/cofunclabs/cofunc/service"
	"github.com/cofunclabs/cofunc/service/exported"
	"github.com/gorilla/mux"
)

type FlowListHandler struct {
	svc *service.SVC
}

func (h *FlowListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

//
// /v1/flows/add?filename=xxx&md5=xxx
type FlowAddHandler struct {
	svc *service.SVC
}

func (h *FlowAddHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	md5 := r.URL.Query().Get("md5")

	// TODO: save file
	_ = filename
	_ = md5

	// TODO: check md5

	// TODO: parse flowl
}

type FlowReadyHandler struct {
	svc *service.SVC
}

func (h *FlowReadyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

type FlowStartHandler struct {
	svc *service.SVC
}

func (h *FlowStartHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

type FlowStopHandler struct {
	svc *service.SVC
}

func (h *FlowStopHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

type FlowStatusHandler struct {
	svc *service.SVC
}

func (h *FlowStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	fid := nameid.WrapID(id)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var resp service.Writer
	if insight, err := h.svc.InsightFlow(ctx, fid); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp = exported.SimpleError{
			Error: err.Error(),
			Desc:  []string{fid.Value()},
		}
	} else {
		w.WriteHeader(http.StatusOK)
		resp = insight
	}
	if err := resp.JsonWrite(w); err != nil {
		log.Fatalln(err)
	}
}

type FlowRunHandler struct {
	svc *service.SVC
}

func (h *FlowRunHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")

	fid := nameid.New(filename)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var resp service.Writer
	if err := h.svc.RunFlow(ctx, fid, r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp = exported.SimpleError{
			Error: err.Error(),
			Desc:  []string{filename, fid.Value()},
		}
	} else {
		w.WriteHeader(http.StatusOK)
		resp = exported.SimpleSucceed{
			Message: "succeed",
			Desc:    []string{filename, fid.Value()},
		}
	}
	if err := resp.JsonWrite(w); err != nil {
		log.Fatalln(err)
	}
}

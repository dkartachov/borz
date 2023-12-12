package borzlet

import (
	"encoding/json"
	"net/http"
)

func (a *API) getPodsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO implement
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("get pods")
}

func (a *API) createPodHandler(w http.ResponseWriter, r *http.Request) {
	// TODO implement
	w.WriteHeader(http.StatusOK)
}

func (a *API) deletePodHandler(w http.ResponseWriter, r *http.Request) {
	// TODO implement
	w.WriteHeader(http.StatusNoContent)
}

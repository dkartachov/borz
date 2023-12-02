package pod

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dkartachov/borz/internal/model"
	"github.com/go-chi/chi/v5"
)

func (a *API) getPodsHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	var pods []model.Pod

	if name != "" {
		pods = a.borzlet.Store.GetPodsByName(name)
	} else {
		pods = a.borzlet.Store.GetPods()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pods)
}

func (a *API) createPodHandler(w http.ResponseWriter, r *http.Request) {
	var p model.Pod

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&p); err != nil {
		log.Print("error decoding pod", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	p.State = model.Scheduled
	err := a.borzlet.EnqueuePod(p)
	if err != nil {
		http.Error(w, fmt.Sprintf("error queuing pod: %v", err), http.StatusUnprocessableEntity)
		return
	}

	a.borzlet.Store.AddPod(p)
	w.WriteHeader(http.StatusOK)
}

func (a *API) deletePodHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "provide pod name", http.StatusBadRequest)
		return
	}

	pod := a.borzlet.Store.GetPod(name)
	pod.State = model.Stopping

	err := a.borzlet.EnqueuePod(pod)
	if err != nil {
		http.Error(w, fmt.Sprintf("error queuing pod: %v", err), http.StatusUnprocessableEntity)
		return
	}

	a.borzlet.Store.AddPod(pod)
	w.WriteHeader(http.StatusNoContent)
}

package pod

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dkartachov/borz/internal/model"
	"github.com/go-chi/chi/v5"
)

func (a *API) getPodsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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
	a.borzlet.Pods[p.Name] = p
	a.borzlet.EnqueuePod(p)
	w.WriteHeader(http.StatusOK)
}

func (a *API) deletePodHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "provide pod name", http.StatusBadRequest)
		return
	}

	pod := a.borzlet.Pods[name]
	pod.State = model.Stopping
	a.borzlet.Pods[name] = pod
	a.borzlet.EnqueuePod(pod)

	w.WriteHeader(http.StatusNoContent)
}

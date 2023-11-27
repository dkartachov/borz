package pod

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dkartachov/borz/internal/model"
	"github.com/go-chi/chi/v5"
)

func (a *PodAPI) createPodHandler(w http.ResponseWriter, r *http.Request) {
	pod := model.Pod{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&pod)
	if err != nil {
		http.Error(w, "invalid pod request body", http.StatusBadRequest)
		return
	}

	pod.State = model.Pending
	err = a.scheduler.EnqueuePod(pod)
	if err != nil {
		// CHECKME is this the proper status code to respond with?
		http.Error(w, fmt.Sprintf("error queuing pod: %v", err), http.StatusUnprocessableEntity)
		return
	}

	a.database.AddPod(pod)

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusProcessing)
	json.NewEncoder(w).Encode(pod)
}

func (a *PodAPI) deletePodHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "provide pod name", http.StatusBadRequest)
		return
	}

	_, ok := a.database.GetWorkerFromPod(name)
	if !ok {
		http.Error(w, "pod not found", http.StatusNotFound)
		return
	}

	statusCode, body := a.podController.SendPodForDeletion(name)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(body)
}

func (a *PodAPI) getPodsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a.database.GetPods())
}

package pod

import (
	"encoding/json"
	"net/http"

	"github.com/dkartachov/borz/internal/model"
	"github.com/go-chi/chi/v5"
)

func (a *API) createPodHandler(w http.ResponseWriter, r *http.Request) {
	pod := model.Pod{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&pod)
	if err != nil {
		http.Error(w, "invalid pod request body", http.StatusBadRequest)
		return
	}

	pod.State = model.Pending
	a.database.AddPod(pod)
	a.scheduler.EnqueuePod(pod)

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pod)
}

func (a *API) deletePodHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "provide pod name", http.StatusBadRequest)
		return
	}

	_, ok := a.scheduler.PodNameByWorker[name]
	if !ok {
		http.Error(w, "pod not found", http.StatusNotFound)
		return
	}

	statusCode, body := a.scheduler.SendPodForDeletion(name)
	// if statusCode == http.StatusNoContent {
	// 	pod := a.database.GetPod(name)
	// 	pod.State = model.Stopping
	// 	a.database.AddPod(pod)
	// }

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(body)
}

func (a *API) getPodsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a.database.GetPods())
}

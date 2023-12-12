package apiserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dkartachov/borz/pkg/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *Server) createPodHandler(w http.ResponseWriter, r *http.Request) {
	pod := model.Pod{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&pod)
	if err != nil {
		http.Error(w, "invalid pod request body", http.StatusBadRequest)
		return
	}

	pod.ID = uuid.New()
	pod.State = model.Pending

	err = s.store.AddPod(pod)
	if err != nil {
		log.Fatalf("error adding pod to store %s", pod.Name)
		http.Error(w, "error creating pod", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusProcessing)
	json.NewEncoder(w).Encode(pod)
}

func (s *Server) deletePodHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing uuid", http.StatusBadRequest)
		return
	}

	// TODO
	// _, ok := a.database.GetWorkerFromPod(name)
	// if !ok {
	// 	http.Error(w, "pod not found", http.StatusNotFound)
	// 	return
	// }

	// statusCode, body := a.podController.SendPodForDeletion(name)
	// w.WriteHeader(statusCode)
	// json.NewEncoder(w).Encode(body)
}

func (s *Server) getPodsHandler(w http.ResponseWriter, r *http.Request) {
	pods, err := s.store.GetPods()
	if err != nil {
		http.Error(w, "error getting pods", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pods)
}

func (s *Server) getPodHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing uuid", http.StatusBadRequest)
		return
	}

	uuid, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid uuid %s", id), http.StatusBadRequest)
	}

	pod, err := s.store.GetPod(uuid)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting pod %s", uuid), http.StatusInternalServerError)
		return
	}

	if pod == nil {
		http.Error(w, fmt.Sprintf("pod %s not found", uuid), http.StatusNotFound)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pod)
}

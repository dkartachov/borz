package apiserver

import (
	"encoding/json"
	"net/http"
)

func (s *Server) getWorkersHandler(w http.ResponseWriter, r *http.Request) {
	workers, err := s.store.GetWorkers()
	if err != nil {
		http.Error(w, "error getting workers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(workers)
}

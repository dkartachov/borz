package worker

import (
	"encoding/json"
	"net/http"
)

func (a *API) getWorkersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a.database.Workers)
}

package deployment

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dkartachov/borz/internal/model"
	"github.com/go-chi/chi/v5"
)

func (a *API) getDeploymentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a.database.GetDeployments())
}

func (a *API) getDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	// TODO implement handler to get one deployment
	w.Write([]byte("get one deployment"))
}

func (a *API) createDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	deployment := model.Deployment{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&deployment); err != nil {
		log.Print("error decoding deployment", err)
		http.Error(w, "invalid deployment body", http.StatusBadRequest)
		return
	}

	a.database.AddDeployment(deployment)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deployment)
}

func (a *API) deleteDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "no name provided", http.StatusBadRequest)
		return
	}

	// TODO call the deployment controller to stop all pods and remove deployment

	w.WriteHeader(http.StatusNoContent)
}

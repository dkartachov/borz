package deployment

import (
	"github.com/dkartachov/borz/internal/manager/database"
	"github.com/go-chi/chi/v5"
)

type API struct {
	router   *chi.Mux
	database *database.Database
}

func Router(d *database.Database) *chi.Mux {
	a := API{router: chi.NewRouter(), database: d}
	a.router.Get("/", a.getDeploymentsHandler)
	a.router.Get("/{ID}", a.getDeploymentHandler)
	a.router.Post("/", a.createDeploymentHandler)
	a.router.Delete("/{name}", a.deleteDeploymentHandler)

	return a.router
}

package worker

import (
	"github.com/dkartachov/borz/internal/manager/database"
	"github.com/go-chi/chi/v5"
)

type API struct {
	router   *chi.Mux
	database *database.Database
}

func Router(db *database.Database) *chi.Mux {
	a := API{router: chi.NewRouter(), database: db}
	a.router.Get("/", a.getWorkersHandler)

	return a.router
}

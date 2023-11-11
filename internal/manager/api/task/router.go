package task

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
	a.router.Get("/", a.getTasksHandler)
	a.router.Get("/{ID}", a.getTaskHandler)

	return a.router
}

package pod

import (
	"github.com/dkartachov/borz/internal/manager/database"
	"github.com/dkartachov/borz/internal/manager/scheduler"
	"github.com/go-chi/chi/v5"
)

type API struct {
	router    *chi.Mux
	database  *database.Database
	scheduler *scheduler.Scheduler
}

func Router(d *database.Database, s *scheduler.Scheduler) *chi.Mux {
	a := API{router: chi.NewRouter(), database: d, scheduler: s}
	a.router.Post("/", a.createPodHandler)
	a.router.Get("/", a.getPodsHandler)
	a.router.Delete("/{name}", a.deletePodHandler)

	return a.router
}

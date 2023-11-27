package pod

import (
	"github.com/dkartachov/borz/internal/manager/controller"
	"github.com/dkartachov/borz/internal/manager/database"
	"github.com/dkartachov/borz/internal/manager/scheduler"
	"github.com/go-chi/chi/v5"
)

type PodAPI struct {
	router        *chi.Mux
	database      *database.Database
	scheduler     *scheduler.Scheduler
	podController *controller.PodController
}

func Router(d *database.Database, s *scheduler.Scheduler, pc *controller.PodController) *chi.Mux {
	a := PodAPI{router: chi.NewRouter(), database: d, scheduler: s, podController: pc}
	a.router.Post("/", a.createPodHandler)
	a.router.Get("/", a.getPodsHandler)
	a.router.Delete("/{name}", a.deletePodHandler)

	return a.router
}

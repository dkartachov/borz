package pod

import (
	"github.com/dkartachov/borz/internal/worker/borzlet"
	"github.com/go-chi/chi/v5"
)

type API struct {
	router  *chi.Mux
	borzlet *borzlet.Borzlet
}

func Router(b *borzlet.Borzlet) *chi.Mux {
	a := API{router: chi.NewRouter(), borzlet: b}
	a.router.Get("/", a.getPodsHandler)
	a.router.Post("/", a.createPodHandler)
	a.router.Delete("/{name}", a.deletePodHandler)

	return a.router
}

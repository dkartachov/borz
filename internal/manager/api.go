package manager

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Api struct {
	Address string
	Port    int
	Manager Manager
	router  *chi.Mux
}

func (a *Api) Init() {
	a.router = chi.NewRouter()
	// TODO add middleware that checks if a shutdown request was recently received
	// and if so, respond with proper status code
	a.router.Use(middleware.Logger)
	a.router.Route("/tasks", func(r chi.Router) {
		r.Get("/", a.getTasksHandler)
		r.Post("/", a.startTaskHandler)
		r.Delete("/{taskId}", a.stopTaskHandler)
	})
	a.router.Post("/shutdown", a.shutDownHandler)

	log.Printf("[%s] server listening on port %d", a.Manager.name, a.Port)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", a.Address, a.Port), a.router); err != nil {
		log.Print(err)
	}
}

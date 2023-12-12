package apiserver

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dkartachov/borz/pkg/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	address string // CHECKME needed??
	port    int
	store   store.Store
	online  bool
}

func NewServer(address string, port int, s store.Store) *Server {
	return &Server{
		address: address,
		port:    port,
		store:   s,
	}
}

func (s *Server) Start() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(s.onlineMiddleware)

	router.Route("/pods", func(r chi.Router) {
		r.Get("/", s.getPodsHandler)
		r.Get("/{id}", s.getPodHandler)
		r.Post("/", s.createPodHandler)
		r.Delete("/{name}", s.deletePodHandler)
	})

	router.Route("/workers", func(r chi.Router) {
		r.Get("/", s.getWorkersHandler)
	})

	s.online = true

	server := http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.address, s.port),
		Handler: router,
	}

	log.Printf("API server listening on port %d", s.port)
	log.Fatal(server.ListenAndServe())
}

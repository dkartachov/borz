package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dkartachov/borz/internal/worker/api/pod"
	"github.com/dkartachov/borz/internal/worker/borzlet"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Address string
	Port    int
	Worker  string
	Borzlet *borzlet.Borzlet

	router   *chi.Mux
	shutdown chan struct{}
}

func (s *Server) Start() {
	s.init()
	server := http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.Address, s.Port),
		Handler: s.router,
	}

	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	serverCtx, cancelServerCtx := context.WithCancel(context.Background())

	go func() {
		<-s.shutdown

		log.Printf("[%s] shutting down API server", s.Worker)

		shutdownCtx, cancelShutdownCtx := context.WithTimeout(serverCtx, time.Second*60)
		defer cancelShutdownCtx()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("[%s] shutdown timed out, forcing exit: %v", s.Worker, err)
		}

		cancelServerCtx()
	}()

	log.Printf("[%s] server listening on port %d", s.Worker, s.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}

func (s *Server) init() {
	s.shutdown = make(chan struct{})
	s.router = chi.NewRouter()
	s.router.Use(middleware.Logger)
	s.router.Mount("/pods", pod.Router(s.Borzlet))
	// CHECKME Should this be a DELETE endpoint? Should this be moved somewhere else?
	s.router.Delete("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		err := s.Borzlet.StopPods(context.Background())
		if err != nil {
			log.Printf("error stopping pods: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		close(s.shutdown)
		w.WriteHeader(http.StatusOK)
	})
}

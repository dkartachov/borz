package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dkartachov/borz/internal/manager/api/deployment"
	"github.com/dkartachov/borz/internal/manager/api/pod"
	"github.com/dkartachov/borz/internal/manager/api/task"
	"github.com/dkartachov/borz/internal/manager/api/worker"
	"github.com/dkartachov/borz/internal/manager/database"
	"github.com/dkartachov/borz/internal/manager/scheduler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Address   string
	Port      int
	Scheduler *scheduler.Scheduler
	Database  *database.Database
	Manager   string

	router   *chi.Mux
	shutdown chan struct{}
	online   bool
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

		log.Printf("[%s] shutting down server", s.Manager)

		shutdownCtx, cancelShutdownCtx := context.WithTimeout(serverCtx, time.Second*60)
		defer cancelShutdownCtx()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("[%s] shutdown timed out, forcing exit: %v", s.Manager, err)
		}

		cancelServerCtx()
	}()

	log.Printf("[%s] server listening on port %d", s.Manager, s.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// wait for server to shutdown via cancelServerCtx()
	<-serverCtx.Done()
}

func (s *Server) init() {
	s.shutdown = make(chan struct{})
	s.router = chi.NewRouter()
	// TODO add middleware that checks if a shutdown request was recently received
	// and if so, respond with proper status code
	s.router.Use(middleware.Logger)
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !s.online {
				http.Error(w, "API server offline", http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	s.router.Mount("/tasks", task.Router(s.Database))
	s.router.Mount("/deployments", deployment.Router(s.Database))
	s.router.Mount("/workers", worker.Router(s.Database))
	s.router.Mount("/pods", pod.Router(s.Database, s.Scheduler))
	// CHECKME Should this be a DELETE endpoint? Should this be moved somewhere else?
	s.router.Delete("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		s.online = false

		go func() {
			client := http.Client{}

			// CHECKME send all shutdown requests concurrently using goroutines?
			for _, worker := range s.Database.GetWorkers() {
				req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/shutdown", worker), nil)
				if err != nil {
					log.Printf("error creating shutdown request for worker %s", worker)
					continue
				}

				resp, err := client.Do(req)
				if err != nil {
					log.Printf("error connecting to worker %s", worker)
					continue
				}

				// CHECKME what to do if worker fails to shutdown?
				if resp.StatusCode != http.StatusOK {
					log.Printf("error shutting down worker %s", worker)
				}

				resp.Body.Close()
			}

			close(s.shutdown)
		}()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("graceful shutdown initiated"))
	})

	s.online = true
}

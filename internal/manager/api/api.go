package api

import (
	"fmt"
	"log"
	"net/http"

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
	router    *chi.Mux
	Manager   string
	online    bool
}

func (s *Server) Start() {
	s.init()
	server := http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.Address, s.Port),
		Handler: s.router,
	}

	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	// serverCtx, cancelServerCtx := context.WithCancel(context.Background())

	// go func() {
	// 	<-a.Manager.Signal.ShutdownAPI
	// 	log.Printf("[%s] shutting down API server", a.Manager.Name)
	// 	shutdownCtx, cancelShutdownCtx := context.WithTimeout(serverCtx, time.Second*60)
	// 	defer cancelShutdownCtx()
	// 	if err := server.Shutdown(shutdownCtx); err != nil {
	// 		log.Fatalf("[%s] shutdown timed out, forcing exit: %v", a.Manager.Name, err)
	// 	}
	// 	cancelServerCtx()
	// }()

	log.Printf("[%s] server listening on port %d", s.Manager, s.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// <-serverCtx.Done()
}

func (s *Server) init() {
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

	s.online = true
}

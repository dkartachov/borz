package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dkartachov/borz/internal/manager/api/deployment"
	"github.com/dkartachov/borz/internal/manager/api/task"
	"github.com/dkartachov/borz/internal/manager/api/worker"
	"github.com/dkartachov/borz/internal/manager/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type API struct {
	Address  string
	Port     int
	Database *database.Database
	router   *chi.Mux
	manager  string
	online   bool
}

func (a *API) Start() {
	a.init()
	server := http.Server{
		Addr:    fmt.Sprintf("%s:%d", a.Address, a.Port),
		Handler: a.router,
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

	log.Printf("[%s] server listening on port %d", a.manager, a.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// <-serverCtx.Done()
}

func (a *API) init() {
	a.router = chi.NewRouter()
	// TODO add middleware that checks if a shutdown request was recently received
	// and if so, respond with proper status code
	a.router.Use(middleware.Logger)
	a.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !a.online {
				http.Error(w, "API server offline", http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	a.router.Mount("/tasks", task.Router(a.Database))
	a.router.Mount("/deployments", deployment.Router(a.Database))
	a.router.Mount("/workers", worker.Router(a.Database))

	a.online = true
}

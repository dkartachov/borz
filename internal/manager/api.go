package manager

import (
	"context"
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
	online  bool
}

func (a *Api) Start() {
	a.init()
	server := http.Server{
		Addr:    fmt.Sprintf("%s:%d", a.Address, a.Port),
		Handler: a.router,
	}

	// serverCtx, serverStopCtx := context.WithCancel(context.Background())
	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// go func() {
	// 	<-sig
	// 	shutdownCtx, shutdownStopCtx := context.WithTimeout(serverCtx, 30*time.Second)
	// 	defer shutdownStopCtx()

	// 	go func() {
	// 		<-shutdownCtx.Done()
	// 		if shutdownCtx.Err() == context.DeadlineExceeded {
	// 			log.Fatal("graceful shutdown timed out, forcing exit")
	// 		}
	// 	}()

	// 	// TODO do other stuff like stop tasks
	// 	if err := server.Shutdown(shutdownCtx); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	serverStopCtx()
	// }()

	go func() {
		<-a.Manager.signal.ShutdownAPI
		log.Printf("[%s] shutting down API server", a.Manager.name)
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("[%s] server listening on port %d", a.Manager.name, a.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func (a *Api) init() {
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
	a.router.Route("/tasks", func(r chi.Router) {
		r.Get("/", a.getTasksHandler)
		r.Post("/", a.startTaskHandler)
		r.Delete("/{taskId}", a.stopTaskHandler)
	})
	a.router.Post("/shutdown", a.shutDownHandler)
	a.online = true
}

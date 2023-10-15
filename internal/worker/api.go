package worker

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Api struct {
	Address string
	Port    int
	Worker  Worker
	router  *chi.Mux
}

func (a *Api) Start() {
	a.init()
	server := http.Server{
		Addr:    fmt.Sprintf("%s:%d", a.Address, a.Port),
		Handler: a.router,
	}

	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	serverCtx, cancelServerCtx := context.WithCancel(context.Background())

	go func() {
		<-a.Worker.Signal.ShutdownAPI
		log.Printf("[%s] shutting down API server", a.Worker.name)
		shutdownCtx, cancelShutdownCtx := context.WithTimeout(serverCtx, time.Second*60)
		defer cancelShutdownCtx()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("[%s] shutdown timed out, forcing exit: %v", a.Worker.name, err)
		}
		cancelServerCtx()
	}()

	log.Printf("[%s] server listening on port %d", a.Worker.name, a.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}

func (a *Api) init() {
	a.router = chi.NewRouter()
	a.router.Use(middleware.Logger)
	a.router.Route("/tasks", func(r chi.Router) {
		r.Get("/", a.getTasksHandler)
		r.Post("/", a.addTaskHandler)
	})
	a.router.Post("/shutdown", a.shutDownHandler)
	a.router.Get("/alive", a.aliveHandler)
}

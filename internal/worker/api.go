package worker

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
	Worker  Worker
	router  *chi.Mux
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
		<-a.Worker.Signal.ShutdownAPI
		log.Printf("[%s] shutting down API server", a.Worker.name)
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("[%s] server listening on port %d", a.Worker.name, a.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
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

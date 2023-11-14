package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dkartachov/borz/internal/worker/api/pod"
	"github.com/dkartachov/borz/internal/worker/borzlet"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type API struct {
	Address string
	Port    int
	Worker  string
	Borzlet *borzlet.Borzlet
	router  *chi.Mux
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
	// 	<-a.Worker.Signal.ShutdownAPI
	// 	log.Printf("[%s] shutting down API server", a.Worker.name)
	// 	shutdownCtx, cancelShutdownCtx := context.WithTimeout(serverCtx, time.Second*60)
	// 	defer cancelShutdownCtx()
	// 	if err := server.Shutdown(shutdownCtx); err != nil {
	// 		log.Fatalf("[%s] shutdown timed out, forcing exit: %v", a.Worker.name, err)
	// 	}
	// 	cancelServerCtx()
	// }()

	log.Printf("[%s] server listening on port %d", a.Worker, a.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// <-serverCtx.Done()
}

func (a *API) init() {
	a.router = chi.NewRouter()
	a.router.Use(middleware.Logger)
	a.router.Mount("/pods", pod.Router(a.Borzlet))
}

package borzlet

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	address string
	port    int
	borzlet *Borzlet

	router   *chi.Mux
	shutdown chan struct{}
}

func NewBorzletServer(address string, port int, b *Borzlet) *Server {
	return &Server{
		address:  address,
		port:     port,
		borzlet:  b,
		router:   chi.NewRouter(),
		shutdown: make(chan struct{}),
	}
}

func (s *Server) Start() {
	s.init()
	server := http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.address, s.port),
		Handler: s.router,
	}

	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	serverCtx, cancelServerCtx := context.WithCancel(context.Background())

	go func() {
		<-s.shutdown

		log.Print("shutting down borzlet server")

		shutdownCtx, cancelShutdownCtx := context.WithTimeout(serverCtx, time.Second*60)
		defer cancelShutdownCtx()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("shutdown timed out, forcing exit: %v", err)
		}

		cancelServerCtx()
	}()

	log.Printf("borzlet server listening on port %d", s.port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}

func (s *Server) init() {
	s.router.Use(middleware.Logger)
	s.router.Mount("/pods", Router(s.borzlet))
	// CHECKME Should this be a DELETE endpoint? Should this be moved somewhere else?
	s.router.Delete("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		err := s.borzlet.StopPods(context.Background())
		if err != nil {
			log.Printf("error stopping pods: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		close(s.shutdown)
		w.WriteHeader(http.StatusOK)
	})
}

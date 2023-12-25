package main

import (
	"github.com/dkartachov/borz/pkg/apiserver"
	"github.com/dkartachov/borz/pkg/model"
	"github.com/dkartachov/borz/pkg/scheduler"
	"github.com/dkartachov/borz/pkg/store"
)

func main() {
	store := store.NewMemoryStore()
	store.AddWorker(model.Worker{Addr: "http://localhost:3001"})
	store.AddWorker(model.Worker{Addr: "http://localhost:3002"})

	scheduler := scheduler.NewScheduler("http://localhost:3000")
	go scheduler.Start()

	server := apiserver.NewServer("localhost", 3000, store)
	server.Start()
}

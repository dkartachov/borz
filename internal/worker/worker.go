package worker

import (
	"log"
	"strconv"

	"github.com/dkartachov/borz/internal/worker/api"
	"github.com/dkartachov/borz/internal/worker/borzlet"
	"github.com/golang-collections/collections/queue"
)

func Run(args []string) {
	name := args[0]
	port, _ := strconv.Atoi(args[1])

	s := borzlet.Store{}
	s.Init()

	b := borzlet.Borzlet{
		JobQueue: queue.New(),
		Store:    &s,
	}
	server := api.Server{
		Address: "localhost",
		Port:    port,
		Worker:  name,
		Borzlet: &b,
	}
	go b.Start()
	server.Start()

	log.Printf("[%v] exiting", name)
}

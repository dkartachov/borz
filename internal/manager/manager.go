package manager

import (
	"log"
	"strconv"
	"strings"

	"github.com/dkartachov/borz/internal/manager/api"
	"github.com/dkartachov/borz/internal/manager/database"
	"github.com/dkartachov/borz/internal/manager/scheduler"
	"github.com/golang-collections/collections/queue"
)

func Run(args []string) {
	name := args[0]
	port, _ := strconv.Atoi(args[1])
	workers := strings.Split(args[2], ",")

	db := database.Database{}
	db.Init()
	db.AddWorkers(workers)

	sched := scheduler.Scheduler{
		PodQueue:        queue.New(),
		PodNameByWorker: make(map[string]string),
		NextWorker:      0,
		Database:        &db,
	}

	server := api.Server{
		Manager:   name,
		Address:   "localhost",
		Port:      port,
		Scheduler: &sched,
		Database:  &db,
	}

	go sched.Start()
	server.Start()

	log.Printf("[%s] exiting", name)
}

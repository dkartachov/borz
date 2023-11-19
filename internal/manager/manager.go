package manager

import (
	"log"
	"strconv"
	"strings"

	"github.com/dkartachov/borz/internal/manager/api"
	"github.com/dkartachov/borz/internal/manager/database"
	"github.com/dkartachov/borz/internal/manager/scheduler"
	"github.com/dkartachov/borz/internal/model"
	"github.com/golang-collections/collections/queue"
)

func Run(args []string) {
	name := args[0]
	port, _ := strconv.Atoi(args[1])
	workers := strings.Split(args[2], ",")

	db := database.Database{
		Workers:     workers,
		Deployments: make(map[string]model.Deployment),
		Pods:        make(map[string]model.Pod),
	}
	sched := scheduler.Scheduler{
		PodQueue:        queue.New(),
		PodNameByWorker: make(map[string]string),
		Workers:         workers,
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

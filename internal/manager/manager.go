package manager

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/dkartachov/borz/internal/manager/api"
	"github.com/dkartachov/borz/internal/manager/database"
	"github.com/dkartachov/borz/internal/manager/scheduler"
	"github.com/dkartachov/borz/internal/model"
)

func Run(args []string) {
	name := args[0]
	port, _ := strconv.Atoi(args[1])
	workers := strings.Split(args[2], ",")

	db := database.Database{}
	db.Init()
	db.AddWorkers(workers)

	sched := scheduler.Scheduler{
		// TODO make channel size configurable
		PodQueue:        make(chan model.Pod, 1),
		PodNameByWorker: make(map[string]string),
		NextWorker:      0,
		Database:        &db,
		Client:          &http.Client{},
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

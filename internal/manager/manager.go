package manager

import (
	"log"
	"strconv"
	"strings"

	"github.com/dkartachov/borz/internal/manager/apiserver"
	"github.com/dkartachov/borz/internal/manager/controller"
	"github.com/dkartachov/borz/internal/manager/database"
	"github.com/dkartachov/borz/internal/manager/scheduler"
)

func Run(args []string) {
	name := args[0]
	port, _ := strconv.Atoi(args[1])
	workers := strings.Split(args[2], ",")

	db := &database.Database{}
	db.Init()
	db.AddWorkers(workers)

	scheduler := scheduler.New(db)
	podController := controller.NewPodController(db)
	server := apiserver.New("localhost", port, scheduler, db, podController)

	go scheduler.Start()
	go podController.Start()
	server.Start()

	log.Printf("[%s] exiting", name)
}

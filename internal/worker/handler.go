package worker

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dkartachov/borz/internal/task"
)

func (a *Api) getTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a.Worker.GetTasks())
}

func (a *Api) addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var t task.Task

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&t); err != nil {
		log.Printf("[%v] error decoding json: %v", a.Worker.name, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a.Worker.AddTask(t)
	w.WriteHeader(http.StatusOK)
}

// shutDownHandler handles a graceful shutdown request.
//
// (1) The server is shut down first to prevent the worker from
// receiving new tasks from a manager. (2) Worker's shutdown channel is closed
// thereby sending a signal to the task scheduler to stop scheduling tasks
func (a *Api) shutDownHandler(w http.ResponseWriter, r *http.Request) {
	close(a.Worker.Signal.ShutdownTaskRunner)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("graceful shutdown process started"))
}

package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/dmuth/google-go-log4go"
	"github.com/gorilla/mux"

	"github.com/alde/eremetic/types"
)

var requests = make(chan *types.Request)
var scheduler *eremeticScheduler

// AddTask handles adding a task to the queue
func AddTask(w http.ResponseWriter, r *http.Request) {
	var request types.Request

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	handleError(err, w)

	err = json.Unmarshal(body, &request)
	handleError(err, w)

	taskId, err := scheduleTask(scheduler, request)
	if err != nil {
		writeJSON(500, err, w)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/task/%s", taskId))
	writeJSON(http.StatusAccepted, taskId, w)
}

// GetTaskInfo returns information about the given task.
func GetTaskInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["taskId"]
	log.Debugf("Fetching task for id: %s", id)
	task := runningTasks[id]

	if task == (eremeticTask{}) {
		writeJSON(http.StatusNotFound, nil, w)
		return
	}

	writeJSON(http.StatusOK, task, w)
}

// Run the RequestChannel Listener
func Run() {
	runningTasks = make(map[string]eremeticTask)
	scheduler = createEremeticScheduler()
	driver, err := createDriver(scheduler)

	if err != nil {
		log.Errorf("Unable to create scheduler driver: %s", err)
		return
	}

	defer close(scheduler.shutdown)
	defer driver.Stop(false)

	go func() {
		if status, err := driver.Run(); err != nil {
			log.Errorf("Framework stopped with status %s and error: %s\n", status.String(), err.Error())
		}
		log.Info("Exiting...")
	}()

	log.Debug("Entering handler.Run loop")
	for {
		select {
		case req := <-requests:
			log.Debug("Found a request in the queue!")
			scheduleTask(scheduler, *req)
		}
	}
}

// CleanupTasks is an infinite loop removing terminal Tasks that have stuck around too long.
func CleanupTasks() {
	for {
		for i, t := range runningTasks {
			if t.deleteAt.After(time.Now()) && types.IsTerminalString(t.Status) {
				delete(runningTasks, i)
			}
		}
		time.Sleep(time.Minute * 15)
	}
}

func handleError(err error, w http.ResponseWriter) {
	if err != nil {
		if err = writeJSON(422, err, w); err != nil {
			panic(err)
		}
	}
}

func writeJSON(status int, data interface{}, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

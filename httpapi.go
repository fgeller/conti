package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"

	uuid "github.com/satori/go.uuid"
)

var rxPathTasksRun = regexp.MustCompile(`/tasks/([^/]+)/run`)

type httpAPI struct {
	db          database
	runRequests chan run
}

func newHTTPAPI(db database, runReqs chan run) *httpAPI {
	return &httpAPI{db, runReqs}
}

type createTaskRequest struct {
	Name   string `json:"name"`
	Script string `json:"script"`
}

type readTasksResponse struct {
	Tasks []task `json:"tasks"`
}

func (a *httpAPI) readTasks(w http.ResponseWriter, r *http.Request) {
	var (
		ts  []task
		buf []byte
		err error
	)

	if ts, err = a.db.readTasks(); err != nil {
		log.Printf("failed to read tasks, err=%v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if buf, err = json.Marshal(map[string]interface{}{"tasks": ts}); err != nil {
		log.Printf("failed to marshal tasks, err=%v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	if _, err = w.Write(buf); err != nil {
		log.Printf("failed to write response, err=%v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (a *httpAPI) createTask(w http.ResponseWriter, r *http.Request) {
	var (
		req    createTaskRequest
		buf    []byte
		taskID string
		err    error
	)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("failed to decode payload, err=%v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Name == "" {
		log.Printf("create task request lacks name")
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.Script == "" {
		log.Printf("create task request lacks script")
		http.Error(w, "script is required", http.StatusBadRequest)
		return
	}

	if taskID, err = a.db.createTask(req.Name, req.Script); err != nil {
		log.Printf("failed to create task, err=%v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if buf, err = json.Marshal(map[string]interface{}{"id": taskID}); err != nil {
		log.Printf("failed to marshal tasks, err=%v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	if _, err = w.Write(buf); err != nil {
		log.Printf("failed to write response, err=%v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

}

func (a *httpAPI) findTask(id string) (task, bool) {
	var res task

	tsks, err := a.db.readTasks()
	if err != nil {
		log.Printf("failed to read tasks, err=%v", err)
		return res, false
	}

	for _, res = range tsks {
		if res.ID == id {
			return res, true
		}
	}
	return res, false
}

type runTaskRequest struct {
	TaskID string `json:"task-id"`
	VCRef  string `json:"vc-ref"`
}

func httpFailf(w http.ResponseWriter, status int, msg string, args ...interface{}) {
	log.Printf(msg, args...)
	http.Error(w, http.StatusText(status), status)
}

func (a *httpAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/tasks":
		switch r.Method {
		case http.MethodGet:
			a.readTasks(w, r)
		case http.MethodPost:
			a.createTask(w, r)
		default:
			log.Printf("unsupported method %#v against path %#v", r.Method, r.URL.Path)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

	case rxPathTasksRun.MatchString(r.URL.Path):
		matches := rxPathTasksRun.FindAllStringSubmatch(r.URL.Path, -1)
		if len(matches) != 1 && len(matches[0]) != 2 {
			httpFailf(w, http.StatusNotFound, "unmatched path %#v", r.URL.Path)
			return
		}

		var req runTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpFailf(w, http.StatusBadRequest, "failed to decode body of run request, err=%v", err)
			return
		}

		tsk, ok := a.findTask(req.TaskID)
		if !ok {
			httpFailf(w, http.StatusNotFound, "could not find task for given ID %#v", req.TaskID)
			return
		}

		rn := run{
			ID:     uuid.NewV4().String(),
			TaskID: req.TaskID,
			Script: tsk.Script,
			VCRef:  req.VCRef,
			Status: "pending",
		}

		if err := a.db.writeRun(rn); err != nil {
			httpFailf(w, http.StatusInternalServerError, "could not create run, err=%v", err)
			return
		}

		a.runRequests <- rn

		buf, _ := json.Marshal(map[string]string{"id": rn.ID})
		if _, err := w.Write(buf); err != nil {
			log.Printf("failed to write response err=%v", err)
		}
		return

	default:
		log.Printf("unsupported path %#v", r.URL.Path)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}

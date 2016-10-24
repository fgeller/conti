package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type httpAPI struct {
	db database
}

func newHTTPAPI(db database) *httpAPI {
	return &httpAPI{db}
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
	default:
		log.Printf("unsupported path %#v", r.URL.Path)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}

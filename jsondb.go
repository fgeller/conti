package main

import (
	"encoding/json"
	"io"
	"os"
	"sync"

	uuid "github.com/satori/go.uuid"
)

type jsonDatabase struct {
	sync.RWMutex
	file *os.File

	Tasks []task
}

func newJSONDatabase(fn string) (*jsonDatabase, error) {
	var (
		err error
		res = &jsonDatabase{}
	)

	if res.file, err = os.OpenFile(fn, os.O_CREATE|os.O_RDWR, 0644); err != nil {
		return res, err
	}

	if err = json.NewDecoder(res.file).Decode(res); err == io.EOF {
		return res, res.persist()
	}

	return res, err
}

func (db *jsonDatabase) createTask(name, script string) (string, error) {
	id := uuid.NewV4().String()
	db.Lock()
	defer db.Unlock()
	db.Tasks = append(db.Tasks, task{ID: id, Name: name, Script: script})
	return id, db.persist()
}

func (db *jsonDatabase) persist() error {
	return json.NewEncoder(db.file).Encode(db)
}

func (db *jsonDatabase) readTasks() ([]task, error) {
	db.RLock()
	defer db.RUnlock()
	return db.Tasks, nil
}

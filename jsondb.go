package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"sync"

	uuid "github.com/satori/go.uuid"
)

type jsonDatabase struct {
	sync.RWMutex
	fileName string

	Tasks []task
}

func newJSONDatabase(fn string) (*jsonDatabase, error) {
	var (
		err  error
		file *os.File
		res  = &jsonDatabase{fileName: fn}
	)
	log.Printf("starting json database backed by %#v", fn)

	if file, err = os.OpenFile(fn, os.O_CREATE|os.O_RDWR, 0644); err != nil {
		return res, err
	}
	defer file.Close()

	if err = json.NewDecoder(file).Decode(res); err == io.EOF {
		log.Printf("initializing database.")
		return res, json.NewEncoder(file).Encode(res)
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
	var (
		file *os.File
		err  error
	)

	if file, err = os.OpenFile(db.fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
		return err
	}

	return json.NewEncoder(file).Encode(db)
}

func (db *jsonDatabase) readTasks() ([]task, error) {
	db.RLock()
	defer db.RUnlock()
	return db.Tasks, nil
}

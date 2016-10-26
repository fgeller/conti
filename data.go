package main

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

type database interface {
	createTask(name string, script string) (string, error) // TODO let's pull the id gen out
	writeRun(r run) error
	readTasks() ([]task, error)
}

type task struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Script string `json:"script"`
	Runs   []run  `json:"runs"`
}

type run struct {
	ID     string `json:"id"`
	TaskID string `json:"task-id"`
	VCRef  string `json:"vc-ref"`
	Script string `json:"script"`
	Status string `json:"status"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

type memDatabase struct {
	tasks []task
}

func (db *memDatabase) createTask(name, script string) (string, error) {
	id := uuid.NewV4().String()
	db.tasks = append(db.tasks, task{ID: id, Name: name, Script: script})
	return id, nil
}

func (db *memDatabase) readTasks() ([]task, error) {
	return db.tasks, nil
}

func (db *memDatabase) writeRun(rn run) error {
	for i, t := range db.tasks {
		if t.ID == rn.TaskID {
			for j, r := range t.Runs {
				if rn.ID == r.ID {
					t.Runs[j] = rn
					return nil
				}
			}

			db.tasks[i].Runs = append(db.tasks[i].Runs, rn)
			return nil
		}
	}

	return fmt.Errorf("couldn't find task %#v", rn.TaskID)
}

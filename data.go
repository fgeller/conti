package main

import uuid "github.com/satori/go.uuid"

type database interface {
	createTask(name string, script string) (string, error)
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

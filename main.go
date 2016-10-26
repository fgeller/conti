package main

import "log"

func main() {
	var (
		db      database
		err     error
		runReqs = make(chan run)
	)

	if db, err = newJSONDatabase("/tmp/cd.json"); err != nil {
		log.Fatalf("failed to create database, err=%v", err)
	}

	go newRunner(db, runReqs).start()
	newEndpoint(db, runReqs).start()
}

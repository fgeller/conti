package main

import "log"

func main() {
	var (
		db  database
		err error
	)

	if db, err = newJSONDatabase("/tmp/cd.json"); err != nil {
		log.Fatalf("failed to create database, err=%v", err)
	}

	newEndpoint(db).start()
}

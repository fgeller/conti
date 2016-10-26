package main

import (
	"bytes"
	"log"

	"golang.org/x/crypto/ssh"
)

type runRequest struct{}

type runner struct {
	db   database
	reqs chan run
}

func newRunner(db database, runReqs chan run) *runner {
	return &runner{db, runReqs}
}

func (r *runner) start() {
	log.Printf("runner starts listening")
	for {
		select {
		case req := <-r.reqs:
			log.Printf("runner is picking up %#v", req)
			config := &ssh.ClientConfig{
				User: "root",
				Auth: []ssh.AuthMethod{ssh.Password("root")},
			}
			client, err := ssh.Dial("tcp", "contd-worker.minikube:22", config)
			if err != nil {
				log.Fatal("Failed to dial: ", err)
			}

			session, err := client.NewSession()
			if err != nil {
				log.Fatal("Failed to create session: ", err)
			}
			defer logClose("ssh session to worker", session)

			var outBuf bytes.Buffer
			var errBuf bytes.Buffer
			session.Stdout = &outBuf
			session.Stderr = &errBuf
			if err := session.Run(req.Script); err != nil {
				log.Fatal("Failed to run: " + err.Error())
			}
			req.Stdout = outBuf.String()
			req.Stderr = errBuf.String()

			log.Printf("ran %#v", req)
			if err := r.db.writeRun(req); err != nil {
				log.Fatalf("failed to write run err=%v", err)
			}
		}
	}
}

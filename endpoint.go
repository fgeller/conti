package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type endpoint struct{}

func newEndpoint() *endpoint { return &endpoint{} }

func (ep *endpoint) start() {
	var g githubWebHook
	mux := http.NewServeMux()
	mux.Handle("/github/webhook/", &g)
	addr := "0.0.0.0:7272"

	log.Printf("endpoint starting to listen on %#v", addr)
	log.Fatalf("endpoint died, err=%v", http.ListenAndServe(addr, mux))
}

type githubWebHook struct{}

func (g *githubWebHook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("request %#v with method %#v not allowed.", r.URL.Path, r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var (
		buf     []byte
		payload map[string]interface{}
		out     []byte
		err     error
	)
	if buf, err = ioutil.ReadAll(r.Body); err != nil {
		log.Printf("failed to read body of request %#v err=%v.", r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err = json.Unmarshal(buf, &payload); err != nil {
		log.Printf("failed to unmarshal payload of request %#v err=%v.", r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	out, _ = json.MarshalIndent(payload, "", "  ")
	log.Printf("Received post request against %#v. Payload:\n%s\n", r.URL.Path, out)
}

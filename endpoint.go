package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type endpoint struct{}

func newEndpoint() *endpoint { return &endpoint{} }

func (ep *endpoint) start() {
	var (
		s = newSocket()
		g = newGithubWebHook(s.in)
	)

	go s.start()

	mux := http.NewServeMux()
	mux.Handle("/github/webhook/", g)
	mux.Handle("/socket", s)
	addr := "0.0.0.0:7272"

	log.Printf("endpoint starting to listen on %#v", addr)
	log.Fatalf("endpoint died, err=%v", http.ListenAndServe(addr, mux))
}

type socket struct {
	sync.Mutex
	up    websocket.Upgrader
	in    chan interface{}
	conns []*websocket.Conn
}

func newSocket() *socket {
	return &socket{
		in: make(chan interface{}, 100),
		up: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
	}
}

func (s *socket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		conn *websocket.Conn
		err  error
	)

	log.Printf("received request %#v with headers %#v", r.URL.Path, r.Header)

	if conn, err = s.up.Upgrade(w, r, nil); err != nil {
		log.Printf("failed to upgrade request %#v to websocket, err=%v", r.URL.Path, err)
		return
	}

	s.Lock()
	s.conns = append(s.conns, conn)
	s.Unlock()
	log.Printf("Added socket to relay info to.")
}

func logClose(msg string, c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("failed to close: %v err=%v", msg, err)
	}
}

func (s *socket) start() {
	for {
		select {
		case m := <-s.in:
			s.Lock()
			var next []*websocket.Conn
			for _, c := range s.conns {
				if err := c.WriteJSON(m); err != nil {
					log.Printf("failed to write json to websocket err=%v", err)
					continue
				}
				next = append(next, c)
			}
			log.Printf("relayed update to %v sockets, %v failed.", len(next), len(s.conns)-len(next))
			s.conns = next
			s.Unlock()
		}
	}
}

type githubWebHook struct {
	out chan interface{}
}

func newGithubWebHook(out chan interface{}) *githubWebHook {
	return &githubWebHook{out}
}

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
	log.Printf("received post request against %#v with payload:\n%s\n", r.URL.Path, out)

	g.out <- payload
}

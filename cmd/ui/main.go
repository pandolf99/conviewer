//main package for the ui
//client of conccons
package main

import (
	"net/http"
	"io"
	"os"
	"fmt"
)

func index(w http.ResponseWriter, r *http.Request) {
	//TODO use templates
	//to specify which port should the template listen on
	f, err := os.Open("index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	body, err := io.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

func newServer() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /{$}", index)
	return m
}

func run() error {
	s := newServer()
	if err := http.ListenAndServe("localhost:55555", s); err != nil {
		return err
	}
	//TODO graceful shutdown
	return nil
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}


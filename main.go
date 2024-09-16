package main

import (
	"fmt"
	"net/http"
	"os"
)

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}

func run() error {
	m := http.NewServeMux()
	m.HandleFunc("GET /{$}", index)
	srv := &http.Server{
		Addr: "localhost:5000",
		Handler: m,
	}
	if err := srv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}

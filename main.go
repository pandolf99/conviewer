package main

import (
	"fmt"
	"net/http"
	"os"
	"io"
)

func index(w http.ResponseWriter, r *http.Request) {
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

//global info of the server
type SrvInfo struct {
	NumConnections int
	NumIps int
}

type InfoHandler func(*SrvInfo, http.ResponseWriter, *http.Request)

//use closures to capture info in each handler
func NewInfoHandler(info *SrvInfo, ih InfoHandler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		ih(info, rw, r)
	}
}

func NumConnsHandler(info *SrvInfo, rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "Number of Connections: %d", info.NumConnections)
}

func run() error {
	//will be shared among handlers so be careful
	//with race conditions
	info := new(SrvInfo)
	m := http.NewServeMux()
	m.HandleFunc("GET /{$}", index)
	m.Handle("GET /numConns",
		NewInfoHandler(info, NumConnsHandler))
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

package main

import (
	"fmt"
	"net/http"
	"net"
	"os"
	"io"
	"sync/atomic"
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
	NumConnections int64
	//set of addresses
	Addresses map[string]struct{}
}

type InfoHandler func(*SrvInfo, http.ResponseWriter, *http.Request)

//serve number of connections to the server
func NumConHandler(info *SrvInfo) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "Number of Connections: %d", atomic.LoadInt64(&(info.NumConnections)))
	}
}

//serve adddresses 
func AddrHandler(info *SrvInfo) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		//Reading here
		addr :=  info.Addresses
		addrS := make([]string, len(addr))
		for k := range addr {
			addrS = append(addrS, k)
		}
		fmt.Fprintf(rw, "Addresses: %v", addrS)
	}
}

//callback to use when connections change state
func ConnMetrics(srv *SrvInfo) func(net.Conn, http.ConnState) {
	return func(conn net.Conn, connState http.ConnState) {
		switch connState {
			case http.StateNew:
				atomic.AddInt64(&(srv.NumConnections), 1)
				if _, ok := srv.Addresses[conn.RemoteAddr().String()]; !ok {
					srv.Addresses[conn.RemoteAddr().String()] = struct{}{}
				}
			case http.StateClosed:
				atomic.AddInt64(&(srv.NumConnections), -1)
				delete(srv.Addresses, conn.RemoteAddr().String())
		}
	}
}

func run() error {
	//will be shared among handlers so be careful
	//with race conditions
	info := &SrvInfo{
		NumConnections: 0,
		Addresses: make(map[string]struct{}),
	}
	m := http.NewServeMux()
	m.HandleFunc("GET /{$}", index)
	m.HandleFunc("GET /numConns",
		NumConHandler(info))
	m.HandleFunc("GET /addresses",
		AddrHandler(info))
	srv := &http.Server{
		Addr: "localhost:55555",
		Handler: m,
		ConnState: ConnMetrics(info),
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

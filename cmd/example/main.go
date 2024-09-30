//minimal example on how to use the conviewer
//using only the msgChannel
package main

import (
	"net/http"
	"github.com/pandolf99/conviewer"
	"fmt"
	"time"
	"os"
	"io"
)

func handler(rw http.ResponseWriter, r *http.Request) {
	//wait to simulate work
	time.Sleep(time.Second * 5)
	fmt.Fprint(rw, "Hello World")
}

func newServer() *http.Server {
	m := http.NewServeMux()
	m.HandleFunc("GET /{$}", handler)
	srv := &http.Server{
		Addr: "localhost:55555",
		Handler: m,
	}
	return srv
}


//start the server 
//and print info from the conviewer to 
//output
func run(output io.Writer) error {
	srv := newServer()
	msgChan := make(chan *conviewer.ConMsg)
	//need to listen to msgChan first
	//so that conviewer does not block
	go func() {
		for msg := range msgChan {
			fmt.Fprintf(output, "%+v\n", msg)
		}
	}()
	conviewer.ObserveServer(srv, msgChan)
	if err := srv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func main() {
	err := run(os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

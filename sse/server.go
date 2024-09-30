//provides a simple sse server
//that will listen to notifications
//and send it to subscribed clients
//clients subscribe by sending a GET to /
package sse

import (
	"net/http"
	"fmt"
)

type sseSerializable interface {
	Event() string
	Data() string
}

func Serialize(s sseSerializable) string {
	return fmt.Sprintf(
		"event: %s\ndata: %s\n\n",
		s.Event(), s.Data())
}

type sseClient struct {
	msgChan chan sseSerializable
	//notifies when client closes connection
	//necessary since sse is only 1 sided communication
	unsubChan chan struct{}
}

//main sse handler
//sends updates to subscribed clients
func subscribe(
	subChan chan sseClient,
	unsubChan chan sseClient,
) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			fmt.Println("exiting connection")
		}()
		rw.Header().Set("Content-Type", "text/event-stream")
		rw.Header().Set("Cache-Control", "no-cache")
		rw.Header().Set("Connection", "keep-alive")
		//subscribe this client
		cli := sseClient{
			msgChan: make(chan sseSerializable),
			unsubChan: make(chan struct{}),
		}
		subChan <- cli
		//wait for connection to end
		//only client can gracefully close connection
		//in sse 
		//if server closes abruptly, client will attempt to reconnect
		ctx := r.Context()
		go func() {
			<-ctx.Done()
			unsubChan <- cli
		}()
		for {
			select {
			case msg := <- cli.msgChan:
				fmt.Fprint(rw, Serialize(msg))
				rw.(http.Flusher).Flush()
			//signals client has been succesfully removed
			//from subscribers and can safely exit
			//handler without blocking other routines.
			case <- cli.unsubChan:
				return
			}
		}
	}
}

type sseServer struct {
	//source of notifications
	msgChan chan sseSerializable
	//set of clients
	clients map[sseClient] struct{}
	//channel to unsubscribe clients
	unsubChan chan sseClient
	//channel to subscribe 
	subChan chan sseClient
	//mux to support multiple sse handlers
	//in the future
	mux *http.ServeMux
}

func (srv sseServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	srv.mux.ServeHTTP(rw, r)
}

func (srv sseServer) listen() {
	for {
		select {
			case msg := <- srv.msgChan:
				for c := range srv.clients {
					c.msgChan <- msg
				}
			case c := <- srv.subChan:
				srv.clients[c] = struct{}{}
			case c := <- srv.unsubChan:
				delete(srv.clients, c)
				c.unsubChan <- struct{}{}
		}
	}
}

//new sse server
func NewServer(msgChan chan sseSerializable) sseServer {
	srv := sseServer{
		msgChan: msgChan,
		clients: make(map[sseClient]struct{}),
		unsubChan: make(chan sseClient),
		subChan: make(chan sseClient),
	}
	mux := http.NewServeMux()
	mux.Handle("GET /{$}", subscribe(srv.subChan, srv.unsubChan))
	srv.mux = mux
	go srv.listen()
	return srv
}

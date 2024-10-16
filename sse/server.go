//provides a simple sse server
//that will listen to notifications
//and send it to subscribed clients
//clients subscribe by sending a GET to /
package sse

import (
	"net/http"
	"fmt"
)


//convert ch T to chan SseSerializable
func WrapChan[T SseSerializable](ch chan T) chan SseSerializable {
	ret := make(chan SseSerializable)
	go func() {
		for m := range ch {
			ret <- m
		}
	}()
	return ret
}


//internal representation of a client for the server.
type sseClient struct {
	msgChan chan SseSerializable
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
			msgChan: make(chan SseSerializable),
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
			fmt.Printf("context is done: %s\n", ctx.Err())
			unsubChan <- cli
		}()
		for {
			select {
			case msg := <- cli.msgChan:
				ser := Serialize(msg)
				_, err := rw.Write(ser)
				if err != nil {
					fmt.Println(err)
				}
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
	msgChan chan SseSerializable
	//set of clients
	clients map[sseClient] struct{}
	//channel to unsubscribe clients
	unsubChan chan sseClient
	//channel to subscribe 
	subChan chan sseClient
	//embed http server
	*http.Server
	Mux *http.ServeMux
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

//Start listening for messages in the msgChan 
//and start serving at address
func (srv sseServer) ListenAndServe() error {
	go srv.listen()
	err := srv.Server.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

//new sse server
//listens to MsgChan for notifications
//serves clients notifications at addr/
func NewServer(msgChan chan SseSerializable, addr string) sseServer {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr: addr,
		Handler: mux,
	}
	sse := sseServer{
		msgChan: msgChan,
		clients: make(map[sseClient]struct{}),
		unsubChan: make(chan sseClient),
		subChan: make(chan sseClient),
		Server: srv,
		Mux: mux,
	}
	//main handler of sse
	//should not register line here
	mux.Handle("GET /coninfo", subscribe(sse.subChan, sse.unsubChan))
	return sse
}

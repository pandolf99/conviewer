package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

//initial load of page
//htmx will take care of populating on load
//this is only called on startup
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



//simple loggger for debugging
func simpleLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(os.Stderr, "Request to %s\n", r.URL.Path)
		fmt.Fprintf(os.Stderr, "From %s\n", r.RemoteAddr)
		h.ServeHTTP(rw, r)
	})
}

func NewServer(
	clientsChan chan client,
	closeChan chan client,
	cm *ConnManager,
) http.Handler {
		m := http.NewServeMux()
		m.HandleFunc("GET /{$}", index)
		m.Handle("GET /coninfo", ConnInfoHandler(clientsChan, closeChan, cm))
		return simpleLogger(m)
}

type conMsg struct {
	addr net.Addr
	connState http.ConnState
}

//Connection Manager
//holds the state of the connections in the server
type ConnManager struct {
	numCons int
	//set of addresses
	activeCons map[string] struct{}
	idleCons map[string] struct{}
}

//Listen to connection changes coming from the notifChan
func (cm *ConnManager) Listen(notifChan chan conMsg, msgChan chan string) {
	for msg := range notifChan {
		addr := msg.addr.String()
		switch msg.connState {
		case http.StateNew:
			cm.numCons++
		case http.StateActive:
			delete(cm.idleCons, addr)
			cm.activeCons[addr] = struct{}{}
		case http.StateIdle:
			delete(cm.activeCons, addr)
			cm.idleCons[addr] = struct{}{}
		case http.StateClosed:
			delete(cm.idleCons, addr)
			delete(cm.activeCons, addr)
			cm.numCons--
		}
		//send update to sse listener
		sseMsg := cm.CreateMsg()
		msgChan <- sseMsg
	}
}

//requests a message of the current state to be sent to 
//the provided channel.
//Useful for when a new client subscribes
//to the conn info handler
//will block if receiver is not ready to read
func (cm *ConnManager) SendMsg(ch chan string) {
	msg := cm.CreateMsg()
	ch <- msg
}

func (cm *ConnManager) CreateMsg() string {
	var msg string
	//TODO use templating for nice htmx
	//Also think of whether I want different events
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Number of Connections: %d <br/>", cm.numCons))
	sb.WriteString("ActiveConnections:")
	for k := range cm.activeCons {
		sb.WriteString(" " + k)
	} 
	sb.WriteString(" <br/>")
	sb.WriteString("IdleConnections:")
	for k := range cm.idleCons {
		sb.WriteString(" " + k)
	} 
	msg = fmt.Sprintf("event: conChange\ndata: %s\n\n", sb.String())
	return msg
}

//client subscribed to the connection notifications
type client struct {
	//message where client will receive messages
	msgChan chan string
	//channel to message client
	//that it has succesfully unsubscribed
	//and can return from the handler
	unsub chan struct{}
}

//msgDispatcher listens to the notification channel
//and sends messages to clients subscribed to the /conInfo handle
//while also taking in new connections
//might want to somehow couple the handler and this function
//Handler is special because it handles and SSE connection
//could do an interface or something like that
func msgDispatcher(
	msgChan chan string,
	clientsChan chan client,
	closeChans chan client,
) {
	activeClients := make(map[client] struct{}) 
	for {
		select {
			case msg := <- msgChan:
				for c := range activeClients {
					c.msgChan <- msg
				}
			case c := <- clientsChan:
				activeClients[c] = struct{}{}
			case c := <- closeChans:
				delete(activeClients, c)
				c.unsub <- struct{}{}
		}
	}
}

func ConnInfoHandler(
	clientChans chan client,
	closeChans chan client,
	cm *ConnManager,
) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			fmt.Println("Exiting connection")
		}()
		rw.Header().Set("Content-Type", "text/event-stream")
		rw.Header().Set("Cache-Control", "no-cache")
		//client that represents this instance of the handler
		cli := client{
			msgChan: make(chan string),
			unsub: make(chan struct{}),
		}
		clientChans <- cli
		//send a message to this channel so this client 
		//is aware that itself just connected
		//might send double but no problem
		//TODO find a better way to do this.
		go cm.SendMsg(cli.msgChan)
		//listen for connection end in routine so that c <- msg does not block
		//on the msgDispatcher in the case where connection is dropped but 
		//client has not been removed from activeClients
		ctx := r.Context()
		go func() {
			<-ctx.Done()
			closeChans <- cli
		}()
		for {
			select {
			case msg := <- cli.msgChan:
				fmt.Fprint(rw, msg)
				rw.(http.Flusher).Flush()
			case <- cli.unsub:
				return
			}
		}
	}
}

//callback to use when connections change state
//will notify the conChan a new connection has occured
func ConnMetrics(notifChan chan conMsg) func(net.Conn, http.ConnState) {
	return func(conn net.Conn, connState http.ConnState) {
		fmt.Println(conn.RemoteAddr().String(), connState.String())
		notifChan <- conMsg{conn.RemoteAddr(), connState}
	}
}

func run() error {
	//clientChan to subscribe new clients to the handler
	clientChan :=  make(chan client)
	closeChan := make(chan client) 
	notifChan := make(chan conMsg)
	msgChan := make(chan string)
	cm := &ConnManager{
		numCons: 0,
		activeCons: make(map[string] struct{}),
		idleCons: make(map[string] struct{}),
	}
	m := NewServer(clientChan, closeChan, cm)
	go cm.Listen(notifChan, msgChan)
	go msgDispatcher(msgChan, clientChan, closeChan)
	srv := &http.Server{
		Addr: "localhost:55555",
		Handler: m,
		ConnState: ConnMetrics(notifChan),
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

package conviewer

import (
	"fmt"
	"net"
	"net/http"
	"os"
)


//simple loggger for debugging
func simpleLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(os.Stderr, "Request to %s\n", r.URL.Path)
		fmt.Fprintf(os.Stderr, "From %s\n", r.RemoteAddr)
		h.ServeHTTP(rw, r)
	})
}

//single connection update
type conUpdateMsg struct {
	addr net.Addr
	connState http.ConnState
}

type ConMsg struct {
	numCons int
	//set of addresses
	activeCons map[string] struct{}
	idleCons map[string] struct{}
}

//Listen to connection changes coming from the notifChan
//keep state of the server here
func listen(notifChan chan conUpdateMsg, msgChan chan *ConMsg) {
	var numCons int
	activeCons := make(map[string] struct{})
	idleCons := make(map[string] struct{})
	for msg := range notifChan {
		addr := msg.addr.String()
		switch msg.connState {
		case http.StateNew:
			numCons++
		case http.StateActive:
			delete(idleCons, addr)
			activeCons[addr] = struct{}{}
		case http.StateIdle:
			delete(activeCons, addr)
			idleCons[addr] = struct{}{}
		case http.StateClosed:
			delete(idleCons, addr)
			delete(activeCons, addr)
			numCons--
		}
		//send update to sse listener
		sseMsg := &ConMsg{
			numCons: numCons,
			activeCons: activeCons,
			idleCons: idleCons,
		}
		//will block until MsgChan is ready to read
		//user should start a go routine to read from the channel
		//before firing up the observer
		//And if this blocks then ConnMetrics will block
		//Which is problematic
		//but how to enforce it or make is safe?
		//Maybe have somekind of default reader
		//that gets overriden when a user starts listening.
		msgChan <- sseMsg
	}
}

//func (cm *ConnManager) CreateMsg() string {
	//var msg string
	////TODO use templating for nice htmx
	////Also think of whether I want different events
	//var sb strings.Builder
	//sb.WriteString(fmt.Sprintf("Number of Connections: %d <br/>", cm.numCons))
	//sb.WriteString("ActiveConnections:")
	//for k := range cm.activeCons {
		//sb.WriteString(" " + k)
	//} 
	//sb.WriteString(" <br/>")
	//sb.WriteString("IdleConnections:")
	//for k := range cm.idleCons {
		//sb.WriteString(" " + k)
	//} 
	//msg = fmt.Sprintf("event: conChange\ndata: %s\n\n", sb.String())
	//return msg
//}


//callback to use when connections change state
//will notify the conChan a new connection has occured
func ConnMetrics(notifChan chan conUpdateMsg) func(net.Conn, http.ConnState) {
	return func(conn net.Conn, connState http.ConnState) {
		//wrap this in a go routine to be non blocking?
		//If I do that, the connection notifications might be out of sync.
		//cause multiple go routines will be wanting to write to the channel,
		//and it will be up to the scheduler.
		fmt.Println(conn.RemoteAddr().String(), connState.String())
		notifChan <- conUpdateMsg{conn.RemoteAddr(), connState}
	}
}


//Server middleware to observe connections to the server
//notifications will be sent to the passed in channel
func ObserveServer(srv *http.Server, msgChan chan *ConMsg) {
	notifChan := make(chan conUpdateMsg)
	srv.ConnState = ConnMetrics(notifChan)
	//clientChan to subscribe new clients to the handler
	go listen(notifChan, msgChan)
}



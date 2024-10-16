package conviewer

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
)

// simple loggger for debugging
func simpleLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(os.Stderr, "Request to %s\n", r.URL.Path)
		fmt.Fprintf(os.Stderr, "From %s\n", r.RemoteAddr)
		h.ServeHTTP(rw, r)
	})
}

// single connection update
type conUpdateMsg struct {
	addr      net.Addr
	connState http.ConnState
}

type ConMsg struct {
	NumCons int
	//set of addresses
	ActiveCons map[string]struct{}
	IdleCons   map[string]struct{}
}

// use a custom marshaler to turn the map into an array
// could do it from scratch to reduce cost of making
// the slices. But good for now.
func (msg *ConMsg) MarshalJSON() ([]byte, error) {
	activeCons := make([]string, 0, len(msg.ActiveCons))
	for k := range msg.ActiveCons {
		activeCons = append(activeCons, k)
	}
	idleCons := make([]string, 0, len(msg.IdleCons))
	for k := range msg.IdleCons {
		idleCons = append(idleCons, k)
	}
	s := struct {
		NumCons    int
		ActiveCons []string
		IdleCons   []string
	}{
		NumCons:    msg.NumCons,
		ActiveCons: activeCons,
		IdleCons:   idleCons,
	}
	return json.Marshal(s)
}

func (msg *ConMsg) UnmarshalJSON(data []byte) error {
	fmt.Printf("%s\n", data)
	s := &struct{
		NumCons    int
		ActiveCons []string
		IdleCons   []string
	}{}
	err := json.Unmarshal(data, s)
	if err != nil {
		return err
	}
	msg.NumCons = s.NumCons
	for _, v := range s.ActiveCons {
		msg.ActiveCons[v] = struct{}{}
	}
	for _, v := range s.IdleCons {
		msg.IdleCons[v] = struct{}{}
	}
	return nil
}

//implement sse.SseSerializable
func (msg *ConMsg) Event() []byte {
	return []byte("conUpdate")
}

//default is to produce JSON
func (msg *ConMsg) Data() []byte {
	//no need to check for error as
	//we control the struct we are marshalling
	s, _ := json.Marshal(msg)
	return s
}

//Listen to connection changes coming from the notifChan
//keep state of the server here
func listen(notifChan chan conUpdateMsg, msgChan chan *ConMsg) {
	var numCons int
	activeCons := make(map[string]struct{})
	idleCons := make(map[string]struct{})
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
			NumCons:    numCons,
			ActiveCons: activeCons,
			IdleCons:   idleCons,
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

// callback to use when connections change state
// will notify the conChan a new connection has occured
func connMetrics(notifChan chan conUpdateMsg) func(net.Conn, http.ConnState) {
	return func(conn net.Conn, connState http.ConnState) {
		//wrap this in a go routine to be non blocking?
		//If I do that, the connection notifications might be out of sync.
		//cause multiple go routines will be wanting to write to the channel,
		//and it will be up to the scheduler.
		fmt.Println(conn.RemoteAddr().String(), connState.String())
		notifChan <- conUpdateMsg{conn.RemoteAddr(), connState}
	}
}

// Server middleware to observe connections to the server
// notifications will be sent to the passed in channel
func ObserveServer(srv *http.Server, msgChan chan *ConMsg) {
	notifChan := make(chan conUpdateMsg)
	srv.ConnState = connMetrics(notifChan)
	//clientChan to subscribe new clients to the handler
	listen(notifChan, msgChan)
}

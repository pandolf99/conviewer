package conviewer

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
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

//keeps the state of the connections to the server
type ConObserver struct {
	srv *http.Server
	numCons int
	activeCons map[string]struct{}
	idleCons map[string]struct{}
	msgChan chan *ConMsg
	notifChan chan conUpdateMsg
	lock *sync.RWMutex  
}

//Returns an observer ready to observe the state,
//starts listening to the server.
//Srv must be started before calling the observer.
//Otherwise data races might ensue.
func NewObserver(srv *http.Server, msgChan chan *ConMsg) ConObserver {
	notifChan := make(chan conUpdateMsg)
	lck := new(sync.RWMutex)
	srv.ConnState = connMetrics(notifChan)
	return ConObserver{
		srv: srv,
		numCons: 0,
		activeCons: make(map[string]struct{}),
		idleCons: make(map[string]struct{}),
		msgChan: msgChan,
		notifChan: notifChan,
		lock: lck,
	}
}

//Listen to connection changes coming from the notifChan
func (manager *ConObserver) Listen() {
	for msg := range manager.notifChan {
		manager.lock.Lock()
		addr := msg.addr.String()
		switch msg.connState {
		case http.StateNew:
			manager.numCons++
		case http.StateActive:
			delete(manager.idleCons, addr)
			manager.activeCons[addr] = struct{}{}
		case http.StateIdle:
			delete(manager.activeCons, addr)
			manager.idleCons[addr] = struct{}{}
		case http.StateClosed:
			delete(manager.idleCons, addr)
			delete(manager.activeCons, addr)
			manager.numCons--
		}
		manager.lock.Unlock()
		//send update to sse listener
		manager.lock.RLock()
		sseMsg := &ConMsg{
			NumCons:    manager.numCons,
			ActiveCons: manager.activeCons,
			IdleCons:   manager.idleCons,
		}
		manager.lock.RUnlock()
		//will block until MsgChan is ready to read
		//user should start a go routine to read from the channel
		//before firing up the observer
		//And if this blocks then ConnMetrics will block
		//Which is problematic
		//but how to enforce it or make is safe?
		//Maybe have somekind of default reader
		//that gets overriden when a user starts listening.
		manager.msgChan <- sseMsg
	}
}

//get the state of the server as a ConMsg
//safe to call concurrently while listening
//to updates
func (observer *ConObserver) GetState() ConMsg {
	observer.lock.RLock()
	state := ConMsg{
		NumCons: observer.numCons,
		ActiveCons: observer.activeCons,
		IdleCons: observer.idleCons,
	}
	observer.lock.RUnlock()
	return state
}




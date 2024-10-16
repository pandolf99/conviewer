package main

import (
	"fmt"
	"time"
	"net/http"

	"github.com/pandolf99/conviewer"
	"github.com/pandolf99/conviewer/sse"
)

type sseMsg string

func (m *sseMsg) Event() []byte {
	return []byte("myEvent")
}

func (m *sseMsg) Data() []byte {
	return []byte("hello world")
}

//send dummy messages to chan every
//5 seconds
func sendMsg(ch chan sse.SseSerializable) {
	var s *sseMsg
	for {
		ch <- s
		time.Sleep(time.Second * 5)
	}
}

func main() {
	//create dummy SSE server
	ch := make(chan sse.SseSerializable)
	go sendMsg(ch)
	srv := sse.NewServer(ch, "localhost:55555")   
	fmt.Println("dummy server on localhost:55555")
	go srv.ListenAndServe()
	//create Connection Viewer server
	//and attach ui endpoints
	conCh := make(chan *conviewer.ConMsg)
	fmt.Println("conInfo Server on localhost:55554")
	conSrv := sse.NewServer(sse.WrapChan(conCh), "localhost:55554")
	mux := conSrv.Mux
	//index page
	mux.HandleFunc("GET /ui/", http.HandlerFunc(Index))
	//ui calls this hanlder which in turn calls the
	//main /coninfo handler
	mux.HandleFunc("GET /ui/coninfo", http.HandlerFunc(
		UpdateHandler("localhost:55554/coninfo")))
	//must start the server before attaching the observer
	go conSrv.ListenAndServe()
	//start the observer
	conviewer.ObserveServer(srv.Server, conCh)
}


package main

import (
	"fmt"
	"time"

	"github.com/pandolf99/conviewer"
	"github.com/pandolf99/conviewer/sse"
)

type sseMsg string

func (m sseMsg) Event() string {
	return "myEvent"
}

func (m sseMsg) Data() string {
	return "hello world"
}
	

//send dummy messages to chan every
//5 seconds
func sendMsg(ch chan sse.SseSerializable) {
	var s sseMsg
	for {
		ch <- s
		time.Sleep(time.Second * 5)
	}
}

func main() {
	ch := make(chan sse.SseSerializable)
	go sendMsg(ch)
	srv := sse.NewServer(ch, "localhost:55555")   
	fmt.Println("dummy server on localhost:55555")
	go srv.ListenAndServe()
	conCh := make(chan *conviewer.ConMsg)
	fmt.Println("conInfo Server on localhost:55554")
	conSrv := sse.NewServer(sse.WrapChan(conCh), "localhost:55554")
	//must start listening before starting the observer
	go conSrv.ListenAndServe()
	//start the observer
	conviewer.ObserveServer(srv.Server, conCh)
}

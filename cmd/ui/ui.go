package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"encoding/json"

	"github.com/pandolf99/conviewer"
	"github.com/pandolf99/conviewer/sse"
)

//load main page
func Index(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("index.html")
	defer f.Close()
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

func templateConMsg(conMsg conviewer.ConMsg) []byte {
	var buff bytes.Buffer
	buff.WriteString(fmt.Sprintf("Number of Connections: %d <br>", conMsg.NumCons))
	buff.WriteString(fmt.Sprintf("Active Connections: %s <br>", conMsg.ActiveCons))
	buff.WriteString(fmt.Sprintf("Idle Connections: %s", conMsg.IdleCons))
	return buff.Bytes()
}

func templateConInfo(conMsg conviewer.ConMsg) []byte {
	//hacky templating
	//but ok for simple example
	var buff bytes.Buffer 
	s := `<h1
		hx-ext="sse"
		sse-connect="/ui/coninfo"
		sse-swap="conUpdate"
		hx-trigger="load">`
	buff.WriteString(s)
	buff.Write(templateConMsg(conMsg))
	buff.WriteString("</h1>")
	return buff.Bytes()
}

/*
*/

//extract conMsg from the sseMsg
func extractMsg(msg *sse.SseMsg) (conviewer.ConMsg, error) {
	conMsg := new(conviewer.ConMsg)
	conMsg.ActiveCons = make(map[string]struct{})
	conMsg.IdleCons = make(map[string]struct{})
	err := json.Unmarshal(msg.Data(), conMsg)
	if err != nil {
		return *conMsg, err
	}
	return *conMsg, nil
}

//call here for updates
//this becomes the client for the sse server
//take in the address of the sse server to call it
//sort of a middle man between sse and front end
//this front end is server side rendered so 
//htmx formatting should occur here.
//you could call the subscribe handler directly if you
//are building your own UI
func UpdateHandler(addr string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var c sse.Client
		ch, err := c.SubscribeWithContext(addr, r.Context()) 
		if err != nil {
			http.Error(rw,
			fmt.Sprintf("Subscribing to SSE: %s", err),
			http.StatusInternalServerError )
			return 
		}
		for msg := range ch {
			conMsg, err := extractMsg(msg)
			if err != nil {
				fmt.Println(err)
				return
			}
			template := templateConMsg(conMsg)
			msg.SetEvent([]byte("conUpdate"))
			msg.SetData(template)
			ser := sse.Serialize(msg)
			rw.Header().Set("Content-Type", "text/event-stream")
			rw.Header().Set("Cache-Control", "no-cache")
			rw.Header().Set("Connection", "keep-alive")
			rw.Write(ser)
			rw.(http.Flusher).Flush()
		}
	}
}

func GetState(observer conviewer.ConObserver) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		template := templateConInfo(observer.GetState())
		rw.Write(template)
		return
	}
}


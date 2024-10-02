//main package for the ui
//client of conccons
package ui

import (
	"net/http"
	"io"
	"os"
)

//load main page
func Index(w http.ResponseWriter, r *http.Request) {
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
		var c http.Client
		resp, err := c.Get(addr)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		//if err != nil {
			//for 
		//}
	}
}


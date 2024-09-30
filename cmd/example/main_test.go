package main

import (
	"net/http"
	"testing"
	"os"
	"sync"
)

func TestRun(t *testing.T) {
	//start the server
	go run(os.Stdout)
	//create 3 requests to the server
	var wg sync.WaitGroup
	wg.Add(3)
	for range 3 {
		go func() {
			defer wg.Done()
			_, err := http.Get("http://localhost:55555/")
			if err != nil {
				t.Log(err)
				t.Fail()
			}
		}()
	}
	wg.Wait()
}



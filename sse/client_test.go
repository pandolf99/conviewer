package sse
//Test the sse client with a dummy server

import (
	"fmt"
	"bytes"
	"bufio"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"os"
)

func sseHandler(rw http.ResponseWriter, r *http.Request) {
	var sseMsg SseMsg
	sseMsg.SetEvent([]byte("myEvent"))
	sseMsg.SetData([]byte("Hello World"))
	msg := Serialize(&sseMsg)
	ticker := time.NewTicker(time.Millisecond * 200)
	timer := time.NewTimer(time.Second * 3)
	for {
		select {
			case <-ticker.C:
				rw.Write(msg)
				rw.Write([]byte{'\n', '\n'})
				rw.(http.Flusher).Flush()
			case <-timer.C:
				return
		}
	}
}

func TestClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(sseHandler))
	defer ts.Close()
	var cli Client 
	ch, err := cli.Subscribe(ts.URL)
	t.Log("Made Request")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	for msg := range ch {
		if bytes.Compare(msg.Event(), []byte("myEvent")) != 0 {
			t.Logf("expected event: myEvent\n", )
			t.Logf("actual event: %s\n", msg.Event())
			t.Fail()
		}
		if bytes.Compare(msg.Data(), []byte("Hello World")) != 0 {
			t.Logf("expected data: Hello World\n", )
			t.Logf("actual event: %s\n", msg.Data())
			t.Fail()
		}
		if t.Failed() {
			return
		}
	}
}

func TestScanner(t *testing.T) {
	dirName := "testdata/streamSamples"
	testFiles, err := os.ReadDir(dirName) 
	if err != nil {
		t.Fatal(err)
	}
	for _, dirEntry := range testFiles {
		fName := dirEntry.Name()
		t.Run(fName, func(t *testing.T) {
			f, err := os.ReadFile(fmt.Sprintf("%s/%s", dirName, fName))
			if err != nil {
				t.Fatal(err)
			}
			buf := bytes.NewBuffer(f)
			scanner := bufio.NewScanner(buf)   
			scanner.Split(scanBody)
			for scanner.Scan() {
				tok := scanner.Bytes()
				t.Logf("\n%s", tok)
			}
		})
	}
}

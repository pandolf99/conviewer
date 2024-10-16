package sse

import (
	"fmt"
	//"io"
	"bufio"
	"bytes"
	"context"
	"net/http"
)

type Client struct {
	http.Client
}

//SplitFunc so that the scanner 
//can read one message from the server
//one token is one message delimited 
//by two new lines.
func scanBody(data []byte, atEOF bool) (int, []byte, error) {
	//fmt.Println(data[0])
	//fmt.Println("Scanning data:")
	//fmt.Printf("%s\n", data)
	//fmt.Println("end of data")
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	//find where message ends, delimited by 3 new lines
	if i := bytes.Index(data, []byte{'\n', '\n', '\n'}); i >= 0 {
		return i + 3, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

//read body of response and send through channel
func readtoChan(ch chan *SseMsg, resp *http.Response) {
	defer func() {
		fmt.Println("scanner stopped")
		resp.Body.Close()
		close(ch)
	}()
	//according to SSE spec
	//scanner will stop when server stops writing
	//to response writer
	scanner := bufio.NewScanner(resp.Body)   
	scanner.Split(scanBody)
	for scanner.Scan() {
		tok := scanner.Bytes()
		msg := new(SseMsg)
		//error handle this
		//implies that server is sending not compliant sse
		err := Unserialize(tok, msg)
		if err != nil {
			fmt.Println(err)
		}
		ch <- msg
	}
}

//subscribe to url
//sends the messages to 
//returned chan
func (cli *Client) Subscribe(url string) (chan *SseMsg, error) {
	ch := make(chan *SseMsg)
	resp, err := cli.Get(url)
	if err != nil {
		return nil, err
	}
	go readtoChan(ch, resp)
	return ch, nil
}

//subscribe with a context that will get passed 
//to the request
func (cli *Client) SubscribeWithContext(url string, ctx context.Context) (chan *SseMsg, error) {
	ch := make(chan *SseMsg)
	url = "http://" + url
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Could hot subscribe to server")
	}
	go readtoChan(ch, resp)
	return ch, nil
}

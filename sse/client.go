package sse

import (
	"net/http"
	"bufio"
	"bytes"
)

type Client struct {
	http.Client
}

//SplitFunc so that the scanner 
//can read one message from the server
//one token is one message delimited 
//by two new lines.
func scanBody(data []byte, atEOF bool) (int, []byte, error) {
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

//subscribe to url
//sends the messages to 
//returned chan
func (cli *Client) Subscribe(url string) (chan *SseMsg, error) {
	ch := make(chan *SseMsg)
	resp, err := cli.Get(url)
	if err != nil {
		return nil, err
	}
	go func() {
		//scan tokens
		//according to SSE spec
		defer func() {
			resp.Body.Close()
			//signal to the client user that
			//the server has stopped sending here
			close(ch)
		}()
		scanner := bufio.NewScanner(resp.Body)   
		scanner.Split(scanBody)
		for scanner.Scan() {
			tok := scanner.Bytes()
			msg := new(SseMsg)
			//error handle this
			//implies that server is sending not compliant sse
			Unserialize(tok, msg)
			ch <- msg
		}
	}()
	return ch, nil
}



package sse

import (
	"net/http"
	_ "bufio"
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
	if i := bytes.Index(data, []byte{'\n', '\n'}); i >= 0 {
		return i + 2, data[0:i], nil
	}
	//EOF means server closed connection
	if atEOF {
		return 0, nil, nil
	}
	// Request more data.
	return 0, nil, nil
}

//subscribe to url
//sends the messages to 
//returned chan
//func (cli *Client) Subscribe(url string) (chan SseSerializable, error) {
	//ch := make(chan SseSerializable)
	//resp, err := cli.Get(url)
	//if err != nil {
		//return nil, err
	//}
	//go func() {
		////scan tokens
		////according to SSE spec
		//defer resp.Body.Close()
		//scanner := bufio.NewScanner(resp.Body)   
		//scanner.Split(scanBody)
		//for scanner.Scan() {
			//tok := scanner.Bytes()
		//}
	//}()
	//return ch, nil
//}



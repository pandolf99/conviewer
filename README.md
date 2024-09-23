## ConcCons (Concurrent Connections)

A simple service to view the connections and their states associated with a
server in real time. Server sends [server sent events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events) 
to clients that are subscribed to the 
`/coninfo` endpoint. The information is displayed in a UI that can be viewed
from the browser.

Currently a work in progress.

## Usage: 
Compile the program using the go tool: `go build .` \
Run the program: `./concCons`
Open `localhost:55555` in a browser (open multiple tabs and close them to see
the realtime change).
You can also subscribe to the events directly using curl: `curl -N
localhost:55555/coninfo`

## TODO
- Clean code and architecture
	* Seperate SSE logic from Connection Manager logic
	* Refactor how Connection Manager is passed as a dependency
- Make it so ConcCons can be attached to any Go Server
- Write Tests



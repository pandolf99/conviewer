## Conviewer (Connection Viewer)

A simple library to view the connections and their states associated with a
server in real time. At its simplest, this library provides a Go channel that
recieves notifications about the state of the connections to the server.
Tracking, the remote addresses and whether they are active or idle. See [this
example](https://github.com/pandolf99/conviewer/cmd/example) on how to use it
like this.  \
On top of that, this library provides a minimal [SSE](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events) server that sends
notifications from that channel to all clients that are subscribed to the
`\coninfo` endpoint. An optional HTMX UI is also provided that can act as one of
the clients the SSE server, updating the state of the server in real time.
Examples and documentation on this coming soon.

Currently a work in progress.

## TODO
- Clean code and architecture
	* Make the calls to the connection viewer safer
- Write Tests



## Browser UI
This example shows how to build a UI in htmx that displays the status of the conncections
to an SSE server. 
The SSE server in this example simply sends "hello world" every 5 seconds and is hosted on port `localhost:55555`.
The connection viewer is hosted on `localhost:55554`
The UI (server rendered) is hosted on `localhost:55554/ui`
Note that this is the same server as the connection viewer.
As this is a simple example its okay, but it might not be what you want
for a more complex ui.

## Usage
Run the main executable. Create connections to the dummy SSE server on
`localhost:55555` using curl: `curl -N localhost:55555` or some other client.
Open a browser window on `localhost:55554` and view the connections.

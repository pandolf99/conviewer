## Example
This program creates a server with a simple handler that takes 5 seconds to
respond. It also outputs to standard output the state of the connections to the
server. This is an example of the most minimal way to use this library, that is
by simply using the provided notification channel to print to screen.

## USAGE
Build and run the main executable, and make requests to `http://localhost55555`
using curl or some other client. Or, run `go test` which automatically starts
the server and makes 3 requests.

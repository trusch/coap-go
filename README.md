# Lobaro CoAP GoLang adapter

[Lobaro CoAP](https://github.com/lobaro/lobaro-coap) provides a highly portable CoAP stack for Client and Server running on almost any hardware.

The **GoLang adapter** uses cgo to provide a CoAP stack based on the code of Lobaro CoAP. It can be used for **testing the stack on a PC** and to **write server applications in go** that can handle CoAP connections.

## Getting Started

The project consists of multiple submodules:

* **coap** - A pure Go client library with an API similar to Go's http package. Supports multiple Transports (e.g. RS232).
* **liblobarocoap** - A CGO wrapper around [Lobaro CoAP](https://github.com/lobaro/lobaro-coap) C Implementation.
* **coapmsg** The underlying CoAP message structure used by other packages. Based on [dustin/go-coap](https://github.com/dustin/go-coap).

It is planned to extend the `coap` package to support more transports like UDP, TCP in future. The package will also get some code to setup CoAP servers. First based on `liblobarocoap` and later also in native Go.

Contributions are welcome!

### Prerequisite 
To build the project you need a C compiler and the matching [Go](https://golang.org/dl/) toolkit installed. 

For Windows you can use [MinGW](http://www.mingw.org/) to install the gcc. 

When you have a 32 bit C compiler make sure you also use 32 bit Go. Else cgo will not be able to compile the C code.

### Install the code

```
go get -u github.com/lobaro/coap-go
```

Execute tests
```
go test github.com/lobaro/coap-go
```

To use the library in your project, just import
```
import "github.com/lobaro/coap-go"
```

# Usage

## Observe

```
// Start observing, res is the result of a GET request
res, err := coap.Observe(url)
if err != nil {
	panic(err)
}
// Gracefully cancel the observe at the end of this function
defer coap.CancelObserve(res)
var timeoutCh time.After(60 * time.Second)

for {
	select {
	// res.Next returns a response with a new Next channel
	// as soon as the client receives a notification from the server
	case nextRes, ok := <-res.Next:
		if !ok {
			return
		}
		res = nextRes // update res for next interation
		// Handle the Notification (res)
	case <-time.After(20 * time.Second):
		return // Cancel observe after 20 seconds of silence
	case <-timeoutCh:
		return // Cancel observe after 60 seconds
	}
}
```

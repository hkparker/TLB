TLB
===

A simple Type Length Value protocol implemented with BSON to hand structs between Go applications in an event driven and parallel way.

It follows from [TLJ](https://github.com/hkparker/TLJ), but BSON is much faster.

```
BenchmarkBSON-4           100000             17542 ns/op
BenchmarkJSON-4             1000           2210297 ns/op
```

Concepts
--------

TLB is used to write networked application in Go by expressing the application's behavior in terms of what to do with structs recieved on various sockets.

Here's a rough idea of how TLB came about:

* maybe "*sockets that have a remote certificate I trust are 'trusted' sockets*"
* or "*sockets that send an `Authentication{}` struct with a valid password are 'trusted' sockets*"
* and "*when 'trusted' sockets send a `Message{}`, save it in the database*"
* and also "*when 'trusted' sockets send a `Message{}`, print it*"
* how could this be expressed easily?

Most generally, when *tag* receives *type*, do *func*.  If there are many funcs with the same criteria, run them all in parallel as goroutines.  This library is meant to be used on a variety of networks, from traditional TLS sockets on the internet to anonymity networks such as I2P.

Usage
-----

To use TLB, start by defining some structs you want to pass around.  We want to hold on to references to their types for later.  These structs are just basic examples, anything that can be marshalled to BSON is ok.

```go
type ExampleEvent struct {
	Parameter1	string
	Parameter2	int
}
example_event_inst := reflect.TypeOf(ExampleEvent{})
example_event_ptr := reflect.TypeOf(&ExampleEvent{})

Type ExampleRequest {
	Parameter1	string
}
example_request_inst := reflect.TypeOf(ExampleRequest{})
example_request_ptr := reflect.TypeOf(&ExampleRequest{})

type ExampleResponse {
	Parameter1	string
	Parameter2	string
	Parameter3	string
}
example_response_inst := reflect.TypeOf(ExampleResponse{})
example_response_ptr := reflect.TypeOf(&ExampleResponse{})
```

Then, define Builder functions for each struct that will create and validate the struct from a BSON byte array.  The TLBContext can be used to access the socket that sent this data.  Add these functions to a TypeStore.

```go
func NewExampleEvent(data []byte, context TLBContext) interface{} {
	event := &ExampleEvent{}
	err := bson.Unmarshal(data, &event)
	if err != nil { return nil }
	return event
}

func NewExampleRequest(data []byte, context TLBContext) interface{} {
	request := &ExampleRequest{}
	err := bson.Unmarshal(data, &request)
	if err != nil { return nil }
	return request
}

func NewExampleResponse(data []byte, context TLBContext) interface{} {
	response := &ExampleResponse{}
	err := bson.Unmarshal(data, &response)
	if err != nil { return nil }
	return response
}

type_store := NewTypeStore()
type_store.AddType(example_event_inst, example_event_ptr, NewExampleEvent)
type_store.AddType(example_request_inst, example_event_ptr, NewExampleRequest)
type_store.AddType(example_response_inst, example_event_ptr, NewExampleResponse)
```

A tagging function is used by the server to tag sockets based on their properties.

```go
func TagSocket(socket *net.Conn, server *Server) {
	server.TagSocket(socket, "all")
	// with TLS sockets, a client certificate could be used to tag sockets
	// in I2P, the remote public key could identify sockets
}
```

Next create a Server and a Client that contain the same TypeStore.

```go
listener := // Anything that implements net.UnixListener
server := NewServer(listener, TagSocket, type_store)

socket := // Anything that implement net.Conn
client := NewClient(socket, type_store, false)
```

Hook up some goroutines on the server that run on structs or requests that came from sockets with certain tags.  A type assertion is used to avoid needing reflect to access fields.

```go
server.Accept("all", example_event, func(iface interface{}, context TLBContext) {
	if example_event, ok :=  iface.(*ExampleEvent); ok {
		fmt.Println("a socket tagged \"all\" sent an ExampleEvent struct")
		fmt.Println(example_event.Parameter1)
		fmt.Println(example_event.Parameter2)
	}
})

server.AcceptRequest("all", example_request, func(iface interface{}, context TLBContext) {
	if example_request, ok :=  iface.(*ExampleRequest); ok {
		fmt.Println("a socket tagged \"all\" sent an ExampleRequest request")
		resp := ExampleResponse {
			Parameter1:	"hello",
			Parameter2:	"world",
			Parameter3:	"response",
		}
		context.Respond(resp)
		if err != nil {
			fmt.Println("response did not send")
		}
	}
})
```

It is also possible to insert sockets into an existing server and have them tagged.  This lets peer-to-peer applications dial sockets on startup as well as accept connections once started.

```go
socket := // any net.Conn
server.Insert(socket)
```

Notice how `false` was passed to `NewClient()`.  This put the Client in Client-Server mode, meaning the Client created a goroutine to read data coming back from the server.  This enables stateful requests, but means this socket could not simultaniously be used in a Server.  To put a Client in p2p mode, the third argument to NewClient should be `true`.

```go
// Client-Server mode:
client := NewClient(socket, type_store, false)
// Able to:
client.Message()
req := client.Request()
req.OnResponse()

// P2P mode:
client := NewClient(socket, type_store, true)
// Able to:
server := // a TLB Server
server.Insert(client.Socket)
client.Message()
```

This is what it might look like:

```go
event := ExampleEvent {
	Parameter1:	"test",
	Parameter2:	0,
}
err := client.Message(event)
if err != nil {
	fmt.Println("message did not send")
}

request := ExampleRequest {
	Parameter1:	"test",
}
req, err := client.Request(request)
if err != nil {
	fmt.Println("request did not send")
}
req.OnResponse(example_response, func(iface) {
	if example_response, ok :=  iface.(*ExampleResponse); ok {
		fmt.Println("the request got a response of type ExampleResponse")
		fmt.Println(example_response.Parameter1)
		fmt.Println(example_response.Parameter2)
		fmt.Println(example_response.Parameter3)
	}
})
```

If you only ever want to send one type of struct, create a `StreamWriter` to avoid calling `reflect` every time you send a struct.  This is like a Client in p2p mode that can only send one type of struct.

```go
writer := NewStreamWriter(client, type_store, example_event_inst)
for {
	writer.Write(<-ExampleEventsChan)
}
```

Tests
-----

```
$ go test -race -cover
Running Suite: TLB Suite
========================
Random Seed: 1465096248
Will run 31 of 31 specs

•••••••••••••••••••••••••••••••
Ran 31 of 31 Specs in 1.012 seconds
SUCCESS! -- 31 Passed | 0 Failed | 0 Pending | 0 Skipped PASS
coverage: 92.2% of statements
ok  	github.com/hkparker/TLB	2.040s
```

License
-------

This project is licensed under the MIT license, see LICENSE for more information.

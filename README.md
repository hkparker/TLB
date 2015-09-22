TLJ
===

A simple Type Length Value protocol implemented with JSON to hand structs between Go applications over a variety of networks.

Motivation
----------

I wanted to be able to write networked application in Go by expressing the applications behavior in terms of what to do with structs recieved on various sockets.  This library is meant to be used on a variety of networks, from traditional TLS sockets on the internet to anonymity networks such as I2P.  Aside from the samples in Usage, functionallity also exists in this library to retrieve dead sockets and servers through channels.

Concepts
--------

TLJ contains a server which acts on the unix.Listener interface, and clients that act on the net.Conn interface.  Both server and client reference the same instance of a TypeStore, which holds all the structs that may be passed over the network.  The server also has a tagging function, which tags accepted sockets.

The server can accept structs from sockets with a specific tag with the server.Accept function.  If the server needs to respond, server.AcceptRequest can be used, which provides functionallity to respond down the socket the struct was recieved on.

Clients can use client.Message to send a struct to a server without recieving a response.  If a response, or responses, are desired, clients can use client.Request, which provides functionallity to accept structs in response.

Usage
-----

To use tlj, start by defining some structs you want to pass around.  We want to hold on to references to their types for later.

```
type InformationalEvent struct {
	Parameter1	string
	Parameter2	int
}
informational_event_inst := reflect.TypeOf(InformationalEvent{})
informational_event_ptr := reflect.TypeOf(&InformationalEvent{})

Type InformationRequest {
	Parameter1	string
}
information_request_inst := reflect.TypeOf(InformationRequest{})
information_request_ptr := reflect.TypeOf(&InformationRequest{})

type InformationResponse {
	Parameter1	string
	Parameter2	string
	Parameter3	string
}
information_response_inst := reflect.TypeOf(InformationResponse{})
information_response_ptr := reflect.TypeOf(&InformationResponse{})
```

Then, define funcs for each struct that will create the struct from a JSON byte array.  Add these functions to a TypeStore.

```
func NewInformationalEvent(data []byte) interface{} {
	event := &InformationalEvent{}
	err := json.Unmarshal(data, &event)
	if err != nil { return nil }
	return event
}

func NewInformationRequest(data []byte) interface{} {
	request := &InformationRequest{}
	err := json.Unmarshal(data, &request)
	if err != nil { return nil }
	return request
}

func NewInformationResponse(data []byte) interface{} {
	response := &InformationResponse{}
	err := json.Unmarshal(data, &response)
	if err != nil { return nil }
	return response
}

type_store := NewTypeStore()
type_store.AddType(informational_event_inst, informational_event_ptr, NewInformationalEvent)
type_store.AddType(information_request_inst, informational_event_ptr, NewInformationRequest)
type_store.AddType(information_response_inst, informational_event_ptr, NewInformationResponse)
```

A tagging function is used by the server to tag sockets based on their properties.

```
func TagSocket(socket *net.Conn, server *Server) {
	server.Tags[socket] = append(server.Tags[socket], "all")
	// with TLS sockets, a client certificate could be used to tag sockets
	// in I2P, the remote public key could identify sockets
}
```

Next create a server and a client that contain the same TypeStore.

```
listener := // Anything that implements net.UnixListener
server := NewServer(listener, TagSocket, type_store)

socket := // Anything that implement net.Conn
client := NewClient(socket, type_store)
```

Hook up some goroutines on the server that run on structs or requests that came from sockets with certain tags.  A type assertion is used to avoid needing reflect to access fields.

```
server.Accept("all", informational_event, func(iface interface{}) {
	if informational_event, ok :=  iface.(*InformationalEvent); ok {			// type assertion as builders return an interface{}
		fmt.Println("a socket tagged \"all\" sent an InformationalEvent struct")
		fmt.Println(informational_event.Parameter1)
		fmt.Println(informational_event.Parameter2)
	}
})

server.AcceptRequest("all", information_request, func(iface interface{}, responder Responder) {
	if information_request, ok :=  iface.(*InformationRequest); ok {
		fmt.Println("a socket tagged \"all\" sent an InformationRequest request")
		resp := InformationResponse {
			Parameter1:	"hello",
			Parameter2:	"world",
			Parameter3:	"response",
		}
		err := responder.Respond(resp)
		if err != nil {
			fmt.Println("response did not send")
		}
	}
})
```

From the client side you can send structs as messages, or make requests and hook up goroutines on responses.

```
event := InformationalEvent {
	Parameter1:	"test",
	Parameter2:	0,
}
err := client.Message(event)
if err != nil {
	fmt.Println("message did not send")
}

request := InformationRequest {
	Parameter1:	"test",
}
req, err := client.Request(request)
if err != nil {
	fmt.Println("request did not send")
}
req.OnResponse(information_response, func(iface) {
	if information_response, ok :=  iface.(*InformationResponse); ok {
		fmt.Println("the request got a response of type InformationResponse")
		fmt.Println(information_response.Parameter1)
		fmt.Println(information_response.Parameter2)
		fmt.Println(information_response.Parameter3)
	}
})
```

There can be many calls to server.Accept, server.AcceptRequest, and client.OnResponse with the same conditions but different functions and each will define a goroutine that will be concurrently executed when the condition is met.

For peer-to-peer applications, both sides of a connection may be included in a TLJ server, and only server.Accept and client.Message can be used.  For more traditional client-server applications client.Request and server.AcceptRequest might make more sense.

License
-------

This project is licensed under the MIT license, see LICENSE for more information.

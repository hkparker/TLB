TLJ
===

A simple Type-Length-Value protocol implemented with JSON to hand structs between Go applications over a variety of networks.

Motivation
----------

I wanted to be able to write networked application in Go by expressing the applications behavior in terms of what to do with structs recieved on various sockets.  This library is meant to be used on a variety of networks, from traditional TLS sockets on the internet to anonnymity networks such as I2P.  Aside from the esamples in Usage, functionallity also exists in this library to retrieve dead sockets and servers through channels.

Usage
-----

To use tlj, start by defining some struct you want to pass around.

```
type InformationalEvent struct {
	Parameter1	string
	Parameter2	int
}

Type InformationRequest {
	Parameter1	string
}

type InformationResponse {
	Parameter1	string
	Parameter2	string
	Parameter3	string
}
```

Then, define funcs for each struct that will create the struct from a JSON byte array.  Add these functions to a TypeStore.

```
func NewInformationalEvent{data []byte) interface{} {
	event := &InformationalEvent{}
	err := json.Unmarshal(data, &event)
	if err != nil { return nil }
	return event
}

func NewInformationRequest{data []byte) interface{} {
	request := &InformationRequest{}
	err := json.Unmarshal(data, &request)
	if err != nil { return nil }
	return request
}

func NewInformationResponse{data []byte) interface{} {
	response := &InformationResponse{}
	err := json.Unmarshal(data, &response)
	if err != nil { return nil }
	return response
}

type_store := NewTypeStore()
type_store.AddType(reflect.TypeOf(InformationalEvent{}), NewInformationalEvent)
type_store.AddType(reflect.TypeOf(InformationRequest{}), NewInformationRequest)
type_store.AddType(reflect.TypeOf(InformationResponse{}), NewInformationResponse)
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
server.Accept("all", reflect.TypeOf(InformationalEvent{}), func(iface) {
	if informational_event, ok :=  iface.(*InformationalEvent); ok {			// type assertion as builders return an interface{}
		fmt.Println("a socket tagged \"all\" sent an InformationalEvent struct")
		fmt.Println(informational_event.Parameter1)
		fmt.Println(informational_event.Parameter2)
	}
})

server.AcceptRequest("all", reflect.TypeOf(InformationRequest{}), func(iface, responder) {
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
req.OnResponse(reflect.TypeOf(InformationResponse{}), func(iface) {
	if information_response, ok :=  iface.(*InformationResponse); ok {
		fmt.Println("the request got a response")
		fmt.Println(information_response.Parameter1)
		fmt.Println(information_response.Parameter2)
		fmt.Println(information_response.Parameter3)
	}
})
```

There can be many calls to server.Accept, server.AcceptRequest, and client.OnResponse with the same conditions but different functions and each will define a goroutine that will be concurrently executed when the condition is met.

License
-------

This project is licensed under the MIT license, see LICENSE for more information.

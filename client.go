package tlj

import (
	"fmt"
	"reflect"
	"encoding/json"
	"encoding/binary"
)

type Client struct {
	Socket		*net.Conn
	Types		map[uint16]func()
	TypeCodes	map[reflect.Type]uint16
	Requests	map[uint16]map[reflect.Type][]func(interface{})
	NextID		int
	Writing		*sync.Mutex
}

//move to common
type Request struct {		// maybe call it a query if it is goes to be used both ways?
	RequestID	uint16
	Type		uint16
	Data		string
}


func NewClient(socket *net.Conn, types map[uint16]func(), type_codes map[reflect.Type]uint16) Client {
	client := Client {
		Socket:		socket,
		Types:		types,
		TypeCodes:	type_codes,
		Requests:	make(map[reflect.Type][]func(interface{})),
		NextID:		1
		Writing:	&sync.Mutex{}
	}
    go client.process()
	return client
}

func (client *Client) process() {
	for {
		type_header := make([]byte, 2)
		// read two bytes and validate type
		//n, err := client.Socket.Read(type_header)
		// next if n == 1?
		
		
		size_header := make([]byte, 4)
		// read the size then read that many bytes
		
	}
	// read from client's Socket and decide what to do with the struct
	//read and parse header, read binary of struct
	
	// rather than a struct type and a size, how about a requestID and a size come back.  it is implicit that he stcut that comes back is of response type, a struct that contains the requiest ID, contained stuct type, and struct data.  format it.  this is the same as a request.
	
	// if it is a response to a previous request
	//if msg.type == resp	// they are all going to be of type resp right?  at the outer layer yes.  outer layer is just requestID (and actual inner struct type)
	//	actually parse it
	//	if client.Requests[RequestID] != nil
	//		for each func in the slice of funcs
	//			go func.(msg)
	
}

func (client *Client) getRequestID() {
	id := client.NextID
	client.NextID = id + 1
	return id
}

func (client *Client) Message(instance Interface{}) error {
	message, err := client.format(instance)
	if err != nil { return err }
	client.Writing.Lock()
	_ , err := client.Socket.Write(message)
	client.Writing.Unlock()
	return err
}

func (client *Client) Request(instance Interface{}) Request {
	request := Request {
		RequestID:	client.getRequestID(),
		Type:		client.TypeCodes[Reflect.TypeOf(instance)],
		Data:		json.Marshal(instance)
	}
	Message(request)
	return request
}

func (request *Request) onResponse(struct_type reflect.Type, function func(interface{})) {
	// create location in map if needed?
	request.Client.requests[request.RequestID][struct_type].append(function)
}

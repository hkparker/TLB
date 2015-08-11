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

type Request struct {
	RequestID	uint16
	Type		uint16
	Data		string
	Client		Client
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
	for {	// export this func?  maybe a read next struct?
		type_header := make([]byte, 2)
		// read two bytes and validate type
		//n, err := client.Socket.Read(type_header)
		// next if n == 1?
		
		
		size_header := make([]byte, 4)
		// read the size then read that many bytes
		
	}
	// read from client's Socket and decide what to do with the struct
	//read and parse header, read binary of struct
	

	// requestID, Type, sizeOfData come back, followed by marshalled struct data	(Actually because nested, size is in outer (format) struct)
	
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

func (client *Client) nextResponse() (interface{}, error) {
	
}

func (client *Client) Message(instance interface{}) error {
	message, err := client.format(instance)
	if err != nil { return err }
	client.Writing.Lock()
	_ , err := client.Socket.Write(message)
	client.Writing.Unlock()
	return err
}

func (client *Client) Request(instance interface{}) (Request, error) {
	request := Request {
		RequestID:	client.getRequestID(),
		Type:		client.TypeCodes[Reflect.TypeOf(instance)],
		Data:		json.Marshal(instance),
		Client:		client	// ensure this isn't formatted
	}
	// create a capsule from this request
	err := Message(request)
	return request, err
}

func (request *Request) OnResponse(struct_type reflect.Type, function func(interface{})) {
	//if request.Client.Requests[request.RequestID] == nil {
	//	request.Client.Requests[request.RequestID] = make(map[reflect.Type][]func(interface{}))
	//}
	request.Client.Requests[request.RequestID][struct_type].append(function)
}

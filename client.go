package tlj

import (
	"fmt"
	"reflect"
	"encoding/json"
	"encoding/binary"
)

type Client struct {
	Socket		*net.Conn
	Types		map[uint16]reflect.Type
	TypeCodes	map[reflect.Type]uint16
	Requests	map[uint16]map[reflect.Type][]func(interface{})
	// mutex for writing to Socket
}

type Request struct {
	Client		Client
	RequestID	int
}

func NewClient(socket *net.Conn, types map[reflect.Type]uint16) Client {
	client := Client{
			Socket:		socket,
			TypeCodes:	types,
			Types:		make(map[uint16]reflect.Type),
			Requests:	make(map[reflect.Type][]func(interface{}))
	}
    for k, v := range types{
        client.Types[v] = k
    }
    go process(client)
	// response events
	// start to read from socket and process responses
	return client
}

func process(client Client) {
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
	
	// if it is a response to a previous request
	//if msg.type == resp	// they are all going to be of type resp right?  at the outer layer yes.  outer layer is just requestID (and actual inner struct type)
	//	actually parse it
	//	if client.Requests[RequestID] != nil
	//		for each thing in the slice of funcs
	//		go .(msg)
	
}


func (client *Client) Message(instance Interface{}) error {
	message, err := client.format(instance)
	if err != nil {return err}
	//client.Out <- message	// maybe attempt a mutex lock and a write here to catch a failed write?
	//client.Mutex.Sync{ client.Write(message) }
	return nil
}


func (tljclient *Client) Request(instance Interface{}) Request {
	// generate a random ID for this request, store it in instance
	// write TL-instance to the client's outgoing channel
	// tell the tljclient to expect a struct response with that random ID
}

func (request *Request) onResponse(struct_type reflect.Type, function func(interface{})) {	//reflect.TypeOf(Hayden{}), func(resp interface{})) {
	//request.RequestID0
	new_struct := <- chan_from_process
	for function := range request.Client.requests[request.RequestID][struct_type] {	// check if its all not nil first, and just add, dont execute (happens in process)
		function.(new_struct)
	}
	
}



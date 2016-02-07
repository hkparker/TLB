package tlj

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"sync"
)

//
// A Client is used to wrap a net.Conn interface and send
// TLJ formatted structs through the interface.
//
type Client struct {
	Socket    net.Conn
	TypeStore TypeStore
	Requests  map[uint16]map[uint16][]func(interface{})
	NextID    uint16
	Writing   *sync.Mutex
	Inserting *sync.Mutex
	Dead      chan error
}

//
// Create a new Client with a net.Conn interface and a
// TypeStore containing all types that will be seen on
// the network.
//
func NewClient(socket net.Conn, type_store TypeStore) Client {
	client := Client{
		Socket:    socket,
		TypeStore: type_store,
		Requests:  make(map[uint16]map[uint16][]func(interface{})),
		NextID:    1,
		Writing:   &sync.Mutex{},
		Inserting: &sync.Mutex{},
		Dead:      make(chan error, 1),
	}
	go client.process()
	return client
}

//
// Clients run client.process in a goroutine to read data
// from the server and attempt to parse it as responses to
// previously made requests.
//
func (client *Client) process() {
	context := TLJContext{
		Socket: client.Socket,
	}
	for {
		iface, err := client.TypeStore.NextStruct(client.Socket, context)
		if err != nil {
			client.Dead <- err
			break
		}
		if capsule, ok := iface.(*Capsule); ok {
			recieved_struct := client.TypeStore.BuildType(capsule.Type, []byte(capsule.Data), context)
			if recieved_struct == nil {
				continue
			}
			if client.Requests[capsule.RequestID][capsule.Type] == nil {
				continue
			}
			for _, function := range client.Requests[capsule.RequestID][capsule.Type] {
				go function(recieved_struct)
			}
		}
	}
}

//
// Return the next request ID to use for a capsule and increment
// the counter.
//
func (client *Client) getRequestID() uint16 {
	id := client.NextID
	client.NextID = id + 1
	return id
}

//
// Given any struct in the Client's TypeStore, format the struct
// and write it down the client's net.Conn.
//
func (client *Client) Message(instance interface{}) error {
	message, err := client.TypeStore.Format(instance)
	if err != nil {
		return err
	}
	client.Writing.Lock()
	_, err = client.Socket.Write(message)
	client.Writing.Unlock()
	return err
}

//
// Given any struct in the Client's TypeStore, format the struct
// inside a capsule and write it down the client's net.Conn.
//
func (client *Client) Request(instance interface{}) (Request, error) {
	instance_data, err := json.Marshal(instance)
	if err != nil {
		return Request{}, err
	}
	instance_type, present := client.TypeStore.LookupCode(reflect.TypeOf(instance))
	if !present {
		return Request{}, errors.New("cannot request type not in type stores")
	}
	request := Request{
		RequestID: client.getRequestID(),
		Type:      instance_type,
		Data:      string(instance_data),
		Client:    client,
	}
	capsule := Capsule{
		RequestID: request.RequestID,
		Type:      request.Type,
		Data:      request.Data,
	}
	client.Inserting.Lock()
	client.Requests[request.RequestID] = make(map[uint16][]func(interface{}))
	client.Inserting.Unlock()
	err = client.Message(capsule)
	return request, err
}

//
// A StreamWriter is a Client that can only be used to send one
// type.  Because of this restriction it takes the reflect.Type
// directly and avoids a reflect call on every write.
//
type StreamWriter struct {
	Socket  net.Conn
	TypeID  uint16
	Writing *sync.Mutex
}

//
// Create a new StreamWriter using a net.Conn, a TypeStore, and
// the reflect.Type of the type this StreamWriter should be
// capable of sending.
//
func NewStreamWriter(conn net.Conn, type_store TypeStore, struct_type reflect.Type) (StreamWriter, error) {
	writer := StreamWriter{
		Socket:  conn,
		Writing: &sync.Mutex{},
	}
	type_id, present := type_store.LookupCode(struct_type)
	if !present {
		return writer, errors.New("no type ID in type store")
	}
	writer.TypeID = type_id
	return writer, nil
}

//
// Write a struct using the StreamWriter.
//
func (writer *StreamWriter) Write(obj interface{}) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	type_bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(type_bytes, writer.TypeID)

	length := len(bytes)
	length_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(length_bytes, uint32(length))

	bytes = append(type_bytes, append(length_bytes, bytes...)...)
	writer.Writing.Lock()
	writer.Socket.Write(bytes)
	writer.Writing.Unlock()
	return nil
}

//
// Requests are returned by Clients when stateful requests are made,
// and can be used to handle server responses with request.OnResponse.
//
type Request struct {
	RequestID uint16
	Type      uint16
	Data      string
	Client    *Client
}

//
// OnResponse is used to define the behaviors used to handle responses
// to client.Request.
//
func (request *Request) OnResponse(struct_type reflect.Type, function func(interface{})) {
	if type_id, present := request.Client.TypeStore.LookupCode(struct_type); present {
		request.Client.Inserting.Lock()
		request.Client.Requests[request.RequestID][type_id] = append(request.Client.Requests[request.RequestID][type_id], function)
		request.Client.Inserting.Unlock()
	}
}

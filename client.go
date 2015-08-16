package tlj

import (
	"fmt"
	"reflect"
	"encoding/json"
	"encoding/binary"
)

type Client struct {
	Socket		*net.Conn
	TypeStore	*TypeStore
	Requests	map[uint16]map[uint16][]func(interface{})
	NextID		int
	Writing		*sync.Mutex
	Inserting	*sync.Mutex
	Dead		chan error
}

type Request struct {
	RequestID	uint16
	Type		uint16
	Data		string
	Client		Client
}

func NewClient(socket *net.Conn, type_store *TypeStore) Client {
	client := Client {
		Socket:		socket,
		TypeStore:	type_store,
		Requests:	make(map[uint16]map[uint16][]func(interface{})),
		NextID:		1,
		Writing:	&sync.Mutex{},
		Inserting:	&sync.Mutex{},
		Dead:		make(chan error, 1)
	}
	go client.process()
	return client
}

func (client *Client) process() {
	for {
		capsule, err := nextStruct(client.Socket, client.TypeStore)
		if err != nil {
			client.Dead <- err
			break
		}
		if reflect.TypeOf(capsule) != reflect.TypeOf(Capsule{}) { continue }
		recieved_struct := client.TypeStore.BuildType(capsule.Type, capsule.Data) //b64 decode?
		if recieved_struct == nil { continue }
		if client.Requests[capsule.RequestID][capsule.Type] == nil { continue }
		for function := range(client.Requests[capsule.RequestID][capsule.Type]) {
			go function(recieved_struct)
		}
	}
}

func (client *Client) getRequestID() {
	// cycle over old requests when id reaches max?
	id := client.NextID
	client.NextID = id + 1
	return id
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
		Data:		json.Marshal(instance),	// base64 encode?
		Client:		client
	}
	capsule := Capsule {
		RequestID:	request.RequestID,
		Type:		request.Type,
		Data:		request.Data
	}
	client.Requests[request.RequestID] = make(map[uint16][]func(interface{}))
	err := Message(capsule)
	return request, err
}

func (request *Request) OnResponse(struct_type reflect.Type, function func(interface{})) {
	type_id := request.Client.TypeStore.LookupCode(struct_type)
	if type_id == nil { return }
	request.Client.Inserting.Lock()
	request.Client.Requests[request.RequestID][type_id].append(function)
	request.Client.Inserting.Unlock()
}

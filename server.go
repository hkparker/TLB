package tlj

import (
	"net"
	"reflect"
	"sync"
)

type Server struct {
	Listener	net.Listener
	TypeStore	*TypeStore
	Tag		func(net.Conn, *Server)
	Tags		map[net.Conn][]string
	Sockets		map[string][]net.Conn
	Events		map[string]map[uint16][]func(interface{})
	Requests	map[string]map[uint16][]func(interface{}, Responder)
	FailedServer	chan error
	FailedSockets	chan net.Conn
	TagManipulation	*sync.Mutex
	InsertRequests	*sync.Mutex
	InsertEvents	*sync.Mutex
}

func NewServer(listener net.Listener, tag func(net.Conn, *Server), type_store *TypeStore) Server {
	server := Server {
		Listener:		listener,
		TypeStore:		type_store,
		Tag:			tag,
		Tags:			make(map[net.Conn][]string),
		Sockets:		make(map[string][]net.Conn),
		Events:			make(map[string]map[uint16][]func(interface{})),
		Requests:		make(map[string]map[uint16][]func(interface{}, Responder)),
		FailedServer:		make(chan error, 1),
		FailedSockets:		make(chan net.Conn, 200),
		TagManipulation:	&sync.Mutex{},
		InsertRequests:		&sync.Mutex{},
		InsertEvents:		&sync.Mutex{},
	}
	go server.process()
	return server
}

func (server *Server) Accept(socket_tag string, struct_type reflect.Type, function func(interface{})) {
	if server.Events[socket_tag] == nil {
		server.Events[socket_tag] = make(map[uint16][]func(interface{}))
	}
	type_code, present := server.TypeStore.LookupCode(struct_type)
	if !present { return }
	server.InsertEvents.Lock()
	server.Events[socket_tag][type_code] = append(server.Events[socket_tag][type_code], function)
	server.InsertEvents.Unlock()
}

func (server *Server) AcceptRequest(socket_tag string, struct_type reflect.Type, function func(interface{}, Responder)) {
	if server.Requests[socket_tag] == nil {
		server.Requests[socket_tag] = make(map[uint16][]func(interface{}, Responder))
	}
	type_code, present := server.TypeStore.LookupCode(struct_type)
	if !present { return }
	server.InsertRequests.Lock()
	server.Requests[socket_tag][type_code] = append(server.Requests[socket_tag][type_code], function)
	server.InsertRequests.Unlock()
}

type Responder struct {
	Server		*Server
	Socket		net.Conn
	RequestID	uint16
	WriteLock	sync.Mutex
}

func (responder *Responder) Respond(object interface{}) error {
	response_bytes, err := FormatCapsule(object, responder.Server.TypeStore, responder.RequestID)
	if err != nil { return err }

	responder.WriteLock.Lock()
	_, err = responder.Socket.Write(response_bytes)
	responder.WriteLock.Unlock()
	if err != nil {
		responder.Server.FailedSockets <- responder.Socket
		delete(responder.Server.Tags, responder.Socket)
		return err
	}
	
	return err
}

func (server *Server) process() {
	for {
		socket, err := server.Listener.Accept()
		if err != nil {
			server.FailedServer <- err
			return
		}
		server.Insert(socket)
	}
}

func (server *Server) Insert(socket net.Conn) {
	server.TagManipulation.Lock()
	server.Tag(socket, server)
	server.TagManipulation.Unlock()
	go server.readStructs(socket)
}

func (server *Server) Delete(socket net.Conn) {
	server.TagManipulation.Lock()
	delete(server.Tags, socket)
	for tag, socket_slice := range server.Sockets {
		for i := len(socket_slice) - 1; i >= 0; i-- {
			if socket == socket_slice[i] {
				server.Sockets[tag] = append(socket_slice[:i], socket_slice[i+1:]...)
			}
		}
	}
	server.TagManipulation.Unlock()
}

func (server *Server) readStructs(socket net.Conn) {
	defer socket.Close()
	for {
		obj, err := NextStruct(socket, server.TypeStore)
		if err != nil {
			server.FailedSockets <- socket
			server.Delete(socket)
			return
		}
		tags := server.Tags[socket]
		if obj == nil {
			continue
		} else if reflect.TypeOf(obj) == reflect.TypeOf(&Capsule{}) {
			for _, tag := range(tags) {
				if capsule, ok :=  obj.(*Capsule); ok {
					if server.Requests[tag][capsule.Type] == nil { continue }
					for _, function := range(server.Requests[tag][capsule.Type]) {
						responder := Responder {
							Server:		server,
							Socket:		socket,
							RequestID:	capsule.RequestID,
							WriteLock:	sync.Mutex{},
						}
						recieved_struct := server.TypeStore.BuildType(capsule.Type, []byte(capsule.Data))
						if recieved_struct != nil { go function(recieved_struct, responder) }
					}
				}
			}
		} else {
			for _, tag := range(tags) {
				recieved_type, present := server.TypeStore.LookupCode(reflect.TypeOf(obj))
				if !present { continue }
				if server.Events[tag][recieved_type] == nil { continue }
				for _, function := range(server.Events[tag][recieved_type]) {
					go function(obj)
				}
			}
		}
	}
}

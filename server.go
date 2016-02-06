package tlj

import (
	"net"
	"reflect"
	"sync"
)

type Server struct {
	Listener        net.Listener
	TypeStore       TypeStore
	Tag             func(net.Conn, *Server)
	Tags            map[net.Conn][]string
	Sockets         map[string][]net.Conn
	Events          map[string]map[uint16][]func(interface{}, TLJContext)
	Requests        map[string]map[uint16][]func(interface{}, TLJContext)
	FailedServer    chan error
	FailedSockets   chan net.Conn
	TagManipulation *sync.Mutex
	InsertRequests  *sync.Mutex
	InsertEvents    *sync.Mutex
}

func NewServer(listener net.Listener, tag func(net.Conn, *Server), type_store TypeStore) Server {
	server := Server{
		Listener:        listener,
		TypeStore:       type_store,
		Tag:             tag,
		Tags:            make(map[net.Conn][]string),
		Sockets:         make(map[string][]net.Conn),
		Events:          make(map[string]map[uint16][]func(interface{}, TLJContext)),
		Requests:        make(map[string]map[uint16][]func(interface{}, TLJContext)),
		FailedServer:    make(chan error, 1),
		FailedSockets:   make(chan net.Conn, 200),
		TagManipulation: &sync.Mutex{},
		InsertRequests:  &sync.Mutex{},
		InsertEvents:    &sync.Mutex{},
	}
	go server.process()
	return server
}

func (server *Server) Accept(socket_tag string, struct_type reflect.Type, function func(interface{}, TLJContext)) {
	if server.Events[socket_tag] == nil {
		server.Events[socket_tag] = make(map[uint16][]func(interface{}, TLJContext))
	}
	type_code, present := server.TypeStore.LookupCode(struct_type)
	if !present {
		return
	}
	server.InsertEvents.Lock()
	server.Events[socket_tag][type_code] = append(server.Events[socket_tag][type_code], function)
	server.InsertEvents.Unlock()
}

func (server *Server) AcceptRequest(socket_tag string, struct_type reflect.Type, function func(interface{}, TLJContext)) {
	if server.Requests[socket_tag] == nil {
		server.Requests[socket_tag] = make(map[uint16][]func(interface{}, TLJContext))
	}
	type_code, present := server.TypeStore.LookupCode(struct_type)
	if !present {
		return
	}
	server.InsertRequests.Lock()
	server.Requests[socket_tag][type_code] = append(server.Requests[socket_tag][type_code], function)
	server.InsertRequests.Unlock()
}

func (server *Server) TagSocket(socket net.Conn, tag string) {
	server.Tags[socket] = append(server.Tags[socket], tag)
	server.Sockets[tag] = append(server.Sockets[tag], socket)
}

func ExcludeString(list []string, omit string) []string {
	keep := make([]string, 0)
	for _, val := range list {
		if val != omit {
			keep = append(keep, val)
		}
	}
	return keep
}

func ExcludeConn(list []net.Conn, omit net.Conn) []net.Conn {
	keep := make([]net.Conn, 0)
	for _, val := range list {
		if val != omit {
			keep = append(keep, val)
		}
	}
	return keep
}

func (server *Server) UntagSocket(socket net.Conn, tag string) {
	server.Tags[socket] = ExcludeString(server.Tags[socket], tag)
	server.Sockets[tag] = ExcludeConn(server.Sockets[tag], socket)
	if len(server.Sockets[tag]) == 0 {
		delete(server.Sockets, tag)
	}
	if len(server.Tags[socket]) == 0 {
		delete(server.Tags, socket)
	}
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
	for _, tag := range server.Tags[socket] {
		server.UntagSocket(socket, tag)
	}
	server.TagManipulation.Unlock()
}

func (server *Server) readStructs(socket net.Conn) {
	defer socket.Close()
	context := TLJContext{
		Server: server,
		Socket: socket,
	}
	for {
		obj, err := server.TypeStore.NextStruct(socket, context)
		if err != nil {
			server.FailedSockets <- socket
			server.Delete(socket)
			return
		}
		tags := server.Tags[socket]
		if obj == nil {
			continue
		} else if reflect.TypeOf(obj) == reflect.TypeOf(&Capsule{}) {
			if capsule, ok := obj.(*Capsule); ok {
				for _, tag := range tags {
					if server.Requests[tag][capsule.Type] == nil {
						continue
					}
					for _, function := range server.Requests[tag][capsule.Type] {
						responder := Responder{
							RequestID: capsule.RequestID,
							WriteLock: sync.Mutex{},
						}
						context := TLJContext{
							Server:    server,
							Socket:    socket,
							Responder: responder,
						}
						recieved_struct := server.TypeStore.BuildType(capsule.Type, []byte(capsule.Data), context)
						if recieved_struct != nil {
							go function(recieved_struct, context)
						}
					}
				}
			}
		} else {
			for _, tag := range tags {
				recieved_type, present := server.TypeStore.LookupCode(reflect.TypeOf(obj))
				if !present {
					continue
				}
				if server.Events[tag][recieved_type] == nil {
					continue
				}
				for _, function := range server.Events[tag][recieved_type] {
					context := TLJContext{
						Server: server,
						Socket: socket,
					}
					go function(obj, context)
				}
			}
		}
	}
}

type TLJContext struct {
	Server    *Server
	Socket    net.Conn
	Responder Responder
}

type Responder struct {
	RequestID uint16
	WriteLock sync.Mutex
}

func (context *TLJContext) Respond(object interface{}) error {
	response_bytes, err := context.Server.TypeStore.FormatCapsule(object, context.Responder.RequestID)
	if err != nil {
		return err
	}

	context.Responder.WriteLock.Lock()
	_, err = context.Socket.Write(response_bytes)
	context.Responder.WriteLock.Unlock()
	if err != nil {
		context.Server.FailedSockets <- context.Socket
		context.Server.Delete(context.Socket)
		return err
	}

	return err
}

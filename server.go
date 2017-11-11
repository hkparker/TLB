package tlb

import (
	"net"
	"reflect"
	"sync"
)

//
// A Server wraps a net.Listener and accepts incoming connections,
// tagging them and running any relevant callbacks on valid TLB
// structs received on them.
//
type Server struct {
	Listener        net.Listener
	TypeStore       TypeStore
	Tag             func(net.Conn, *Server)
	Tags            map[net.Conn][]string
	Sockets         map[string][]net.Conn
	Events          map[string]map[uint16][]func(interface{}, TLBContext)
	Requests        map[string]map[uint16][]func(interface{}, TLBContext)
	FailedServer    chan error
	FailedSockets   chan net.Conn
	TagManipulation *sync.Mutex
	InsertRequests  *sync.Mutex
	InsertEvents    *sync.Mutex
}

//
// Create a new server from a net.Listener, a TypeStore, and a tagging
// function that will assign tags to all accepted sockets.
//
func NewServer(listener net.Listener, tag func(net.Conn, *Server), type_store TypeStore) Server {
	server := Server{
		Listener:        listener,
		TypeStore:       type_store,
		Tag:             tag,
		Tags:            make(map[net.Conn][]string),
		Sockets:         make(map[string][]net.Conn),
		Events:          make(map[string]map[uint16][]func(interface{}, TLBContext)),
		Requests:        make(map[string]map[uint16][]func(interface{}, TLBContext)),
		FailedServer:    make(chan error, 1),
		FailedSockets:   make(chan net.Conn, 200),
		TagManipulation: &sync.Mutex{},
		InsertRequests:  &sync.Mutex{},
		InsertEvents:    &sync.Mutex{},
	}
	go server.process()
	return server
}

//
// Create a new callback to be ran when a socket with a certain tag receives
// a specific type of struct.  The server will have no ability to respond
// statefully to this event.
//
func (server *Server) Accept(socket_tag string, struct_type reflect.Type, function func(interface{}, TLBContext)) {
	if server.Events[socket_tag] == nil {
		server.Events[socket_tag] = make(map[uint16][]func(interface{}, TLBContext))
	}
	if type_code, present := server.TypeStore.LookupCode(struct_type); present {
		server.InsertEvents.Lock()
		server.Events[socket_tag][type_code] = append(server.Events[socket_tag][type_code], function)
		server.InsertEvents.Unlock()
	}
}

//
// Create a new callback to be ran when a socket with a certain tag receives
// a capsule containing a specific type of struct.  The callback accepts a
// responder which can be used to respond to the client statefully.
//
func (server *Server) AcceptRequest(socket_tag string, struct_type reflect.Type, function func(interface{}, TLBContext)) {
	if server.Requests[socket_tag] == nil {
		server.Requests[socket_tag] = make(map[uint16][]func(interface{}, TLBContext))
	}
	if type_code, present := server.TypeStore.LookupCode(struct_type); present {
		server.InsertRequests.Lock()
		server.Requests[socket_tag][type_code] = append(server.Requests[socket_tag][type_code], function)
		server.InsertRequests.Unlock()
	}
}

//
// Assign a string tag to a socket in this Server.
//
func (server *Server) TagSocket(socket net.Conn, tag string) {
	server.TagManipulation.Lock()
	server.Tags[socket] = append(server.Tags[socket], tag)
	server.Sockets[tag] = append(server.Sockets[tag], socket)
	server.TagManipulation.Unlock()
}

//
// Given a slice of strings, return all strings that are not equal to omit.
//
func ExcludeString(list []string, omit string) []string {
	keep := make([]string, 0)
	for _, val := range list {
		if val != omit {
			keep = append(keep, val)
		}
	}
	return keep
}

//
// Given a slice of net.Conn interfaces, return all interfaces that are
// not equal to omit.
//
func ExcludeConn(list []net.Conn, omit net.Conn) []net.Conn {
	keep := make([]net.Conn, 0)
	for _, val := range list {
		if val != omit {
			keep = append(keep, val)
		}
	}
	return keep
}

//
// Disassociate a string tag from a socket on this server.
//
func (server *Server) UntagSocket(socket net.Conn, tag string) {
	server.TagManipulation.Lock()
	server.Tags[socket] = ExcludeString(server.Tags[socket], tag)
	server.Sockets[tag] = ExcludeConn(server.Sockets[tag], socket)
	if len(server.Sockets[tag]) == 0 {
		delete(server.Sockets, tag)
	}
	if len(server.Tags[socket]) == 0 {
		delete(server.Tags, socket)
	}
	server.TagManipulation.Unlock()
}

//
// Every Server runs process in a goroutine to accept and Insert new connections.
//
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

//
// Tag the socket then read an structs from this socket until the socket is closed.
//
func (server *Server) Insert(socket net.Conn) {
	server.Tag(socket, server)
	go server.readStructs(socket)
}

//
// Remove all tags from a socket, removing it from the server.
//
func (server *Server) Delete(socket net.Conn) {
	server.TagManipulation.Lock()
	for _, tag := range server.Tags[socket] {
		server.Tags[socket] = ExcludeString(server.Tags[socket], tag)
		server.Sockets[tag] = ExcludeConn(server.Sockets[tag], socket)
		if len(server.Sockets[tag]) == 0 {
			delete(server.Sockets, tag)
		}
		if len(server.Tags[socket]) == 0 {
			delete(server.Tags, socket)
		}
	}
	server.TagManipulation.Unlock()
}

//
// Read structs from a socket until the socket is closed, running any relevant callbacks.
//
func (server *Server) readStructs(socket net.Conn) {
	defer socket.Close()
	context := TLBContext{
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
		server.TagManipulation.Lock()
		tags := server.Tags[socket]
		server.TagManipulation.Unlock()
		if obj == nil {
			continue
		} else if reflect.TypeOf(obj) == reflect.TypeOf(&Capsule{}) {
			server.runRequestCallbacks(obj, tags, context)
		} else {
			server.runEventCallbacks(obj, tags, context)
		}
	}
}

//
// Run all functions stored during server.Accept calls
//
func (server *Server) runEventCallbacks(obj interface{}, tags []string, context TLBContext) {
	for _, tag := range tags {
		recieved_type, present := server.TypeStore.LookupCode(reflect.TypeOf(obj))
		if !present {
			continue
		}
		if server.Events[tag][recieved_type] == nil {
			continue
		}
		for _, function := range server.Events[tag][recieved_type] {
			go function(obj, context)
		}
	}
}

//
// Run all functions stored during server.AcceptRequest calls
//
func (server *Server) runRequestCallbacks(obj interface{}, tags []string, context TLBContext) {
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
				context.Responder = responder
				recieved_struct := server.TypeStore.BuildType(capsule.Type, []byte(capsule.Data), context)
				if recieved_struct != nil {
					go function(recieved_struct, context)
				}
			}
		}
	}
}

//
// Context about TLB events so Server callbacks can respond statefully
// and Builders can conditionally validate data and verify signatures.
//
type TLBContext struct {
	Server    *Server
	Socket    net.Conn
	Responder Responder
}

//
// Responders contain information needed to send a stateful response
//
type Responder struct {
	RequestID uint16
	WriteLock sync.Mutex
}

//
// Respond is used to send a struct down the socket the sent a request
// with client.Request
//
func (context *TLBContext) Respond(object interface{}) error {
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

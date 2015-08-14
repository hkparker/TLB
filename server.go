package tlj

import (
	"fmt"
	"errors"
	"reflect"
	"encoding/json"
	"encoding/binary"
)

type Server struct {
	Listener		*net.UnixListener
	Types			*TypeStore
	Tag				func(*net.Conn)
	Tags			map[*net.Conn][]string
	Events			map[string]map[reflect.Type][]func(interface{})
	Requests		map[string]map[reflect.Type][]func(interface{}, *Responder)
	FailedServer	chan error
	FailedSockets	chan *net.Conn
}

func NewServer(listener *net.UnixListener, tag func(*net.Conn), type_store *TypeStore) Server {
	server := Server {
		Listener:		listener,
		Types:			type_store,
		Tag:			tag,
		Tags:			make(map[*net.Conn][]string),
		Events:			make(map[string]map[reflect.Type][]func(interface{})),
		Requests:		make(map[string]map[reflect.Type][]func(interface{}, *Responder)),
		FailedServer:	make(chan error, 1),
		FailedSockets:	make(chan *net.Conn, 200)
	}
	go server.process()
	return server
}

func (server *Server) Accept(socket_tag string, struct_type reflect.Type, function func(interface{})) {
	// create location in events map if needed?
	server.Events[socket_tag][struct_type] = append(server.Events[socket_tag][struct_type], function)
}

func (server *Server) AcceptRequest(socket_tag string, struct_type reflect.Type, function func(interface{}, responder *Responder)) {
	// create location in requests map if needed?
	server.Requests[socket_tag][struct_type] = append(server.Events[socket_tag][struct_type], function)
}

type Responder struct {
	Server		*Server
	Socket		*net.Conn
	RequestID	uint16
}

func (responder *Responder) Respond(object interface{}) error {
	response_bytes, err := responder.Server.formatCapsule(object, responder.Server.TypeStore, request_id)
	if err != nil { return err }
	
	err = responder.Socket.Write(response_bytes)
	if err != nil {
		responder.Server.FailedSockets <- responder.Socket
		delete(responder.Server.Tags, responder.Socket)
		return err
	}
	
	return nil
}

func (server *Server) process() {
	for {
		socket, err := server.Listener.Accept()
		if err != nil {
			server.FailedServer <- err
			return
		}
		// add to the Sockets map for tagging? or can you append to the slice created by make?
		server.Tag(socket)
		go server.readStructs(socket)
	}
}

func (server *Server) readStructs(socket *net.Conn) {
	for {
		obj, err := nextStruct(socket, server.TypeStore)
		if err != nil {
			server.FailedSockets <- socket
			delete(server.Tags, socket)
			return
		}
		tags := server.Tags[socket]
		if obj == nil {
			continue
		} else if request.TypeOf(obj) == tlj.Capsule {	// or == request.TypeOf(Capsule{})
			for tag := range(tags) { 					// range over nil ok?
				//server.Requests[tag][reflect.TypeOf(obj)]  // make sure this isn't nil?
				for function := range(server.Requests[tag][reflect.TypeOf(obj)]) {
					responder := Responder {
						Server:		server,
						Socket:		socket,
						RequestID:	obj.RequestID
					}
					recieved_struct := server.TypeStore.BuildType(obj.Type, obj.Data)
					if recieved_struct != nil { go function(recieved_struct, responder) }
				}
			}
		} else {
			for tag := range(tags) {
				//server.Events[tag][reflect.TypeOf(obj)]  // make sure this isn't nil?
				for function := range(server.Events[tag][reflect.TypeOf(obj)]) {
					go function(obj)
				}
			}
		}
	}
}

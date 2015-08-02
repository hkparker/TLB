package tlj

import (
	"fmt"
	"reflect"
	"encoding/json"
	"encoding/binary"
)

type Server struct {
	Listener	*net.UnixListener
	Sockets		map[string][]*net.Conn	// other way around?
	Tag			func(*net.Conn)
	Events		map[string]map[reflect.Type][]func(interface{})
	Requests	map[string]map[reflect.Type][]func(*net.Conn, uint16, interface{})
}

func NewServer(listener *net.UnixListener, tag func(*net.Conn)) Server {
	server := Server{
		Listener:	listener,
		Sockets:	make(map[string][]*net.Conn),
		Tag:		tag,
		Events:		make(map[string]map[reflect.Type][]func(interface{})),
		Requests:	make(map[string]map[reflect.Type][]func(*net.Conn, uint16, interface{}))
	}
	
	go server.handleConnections()
	go server.readConnections()
	return server
}

func (server *Server) handleConnections() {
	// conn, err := server.Listener.Accept()	// errors go where?
	// tag the socket
	// server.Connections <- conn
}

func (server *Server) readConnections() {
	// for each thing in the channel
		// go process the structs
}

func (server *Server) readStructs(socket *net.Conn) {
	// if request
		// lookup funcs from Requests
	// if anything else
		// lookup func from Events
}

func (server *Server) Accept(socket_tag string, struct_type reflect.Type, function func(interface{})) {
	// create location in events map if needed?
	server.Events[socket_tag][struct_type] = append(server.Events[socket_tag][struct_type], function)
}

func (server *Server) AcceptRequest(socket_tag string, struct_type reflect.Type, function func(*net.Conn, uint16, interface{})) {
	// create location in requests map if needed?
	server.Requests[socket_tag][struct_type] = append(server.Events[socket_tag][struct_type], function)
}

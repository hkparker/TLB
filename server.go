package tlj

import (
	"fmt"
	"reflect"
	"encoding/json"
	"encoding/binary"
)

type Server struct {
	Listener	*net.UnixListener
	Sockets		map[string][]*net.Conn	// other way around? socket -> slice of tags
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
	
	go server.process()
	return server
}

func (server *Server) process() {
	for {
		socket, err := server.Listener.Accept()
		// break if err?
		server.Tag(socket)
		go server.readStructs(socket)
	}
}

func (server *Server) readStructs(socket *net.Conn) {
	// read struct from socket
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

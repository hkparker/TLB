package tlj

import (
	"fmt"
	"reflect"
	"encoding/json"
	"encoding/binary"
)

type Server struct {
	Listener	*net.UnixListener
	Sockets		map[string][]*net.Conn							// sockets tagged with strings
	Tag			func(*net.Conn)
	Events		map[string]map[reflect.Type][]func(interface{})						//  string tag -> map from types to funcs
	Requests	map[string]map[reflect.Type][]func(*net.Conn, uint16, interface{})	//  string tag -> map from types to funcs
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
	// accept new connections from Listener
	
	// read from all sockets
	// if request
		// lookup funcs from Requests
	// if anything else
		// lookup func from Events
}

func (server *Server) Accept(socket_tag string, struct_type reflect.Type, function func(interface{})) {
	// add to events map
}

func (server *Server) AcceptRequest(socket_tag string, struct_type reflect.Type, function func(*net.Conn, uint16, interface{})) {
	// add to requests map
}

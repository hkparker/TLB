package tlj

import (
	"fmt"
	"reflect"
	"encoding/json"
	"encoding/binary"
)

type Server struct {
	Listener	*net.UnixListener
	Tags		map[*net.Conn][]string
	Types		map[uint16]func()
	TypeCodes	map[reflect.Type]uint16
	Tag			func(*net.Conn)
	Events		map[string]map[reflect.Type][]func(interface{})
	Requests	map[string]map[reflect.Type][]func(*net.Conn, uint16, interface{})
}

type Request struct {
	RequestID	uint16
	Type		uint16
	Data		string
}

type Respose struct {
	RequestID	uint16
	Type		uint16
	Data		string
}

func NewServer(listener *net.UnixListener, tag func(*net.Conn)) Server {
	server := Server{
		Listener:	listener,
		Sockets:	make(map[*net.Conn][]string),
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
		// add to the Sockets map for tagging? or can you append to the slice created by make?
		server.Tag(socket)
		go server.readStructs(socket)
	}
}

func (socket *net.Conn) nextStruct(server Server) (interface{}, []string, error) {
	header := make([]byte, 6)
	_, err := socket.Read(header)
	if err != nil { return nil, nil, err }
	
	type_bytes := header[:2]	// First two bytes are struct type
	size_bytes := header[2:]	// Next four bytes are struct size
	
	type_int := binary.LittleEndian.Uint16(type_bytes)
	size_int := binary.LittleEndian.Uint16(size_bytes)

	struct_data := make([]byte, size_int)
	_, err := socket.Read(struct_data)
	if err != nil { return nil, nil, err }
	
	recieved_struct := server.Types[type_int](struct_data)
	
	tags = server.Tags[socket]
	
	return recieved_struct, tags, nil	
}

func (server *Server) readStructs(socket *net.Conn) {
	for {
		obj, tags, err := socket.nextStruct()
		if err != nil { return }	// signal that this socket closed?  A channel of errored sockets maybe?
		if obj == nil {
			continue
		} else if request.TypeOf(obj) == tlj.Request {	// or == request.TypeOf(Request{})
			for tag := range(tags) {
				//server.Requests[tag][reflect.TypeOf(obj)]  // make sure this isn't nil?
				for function := range(server.Requests[tag][reflect.TypeOf(obj)]) {
					// deconstruct the obj to remove requestID, data..., continue if the inner stuct fails to parse
					requestID 
					go function(socket, requestID, obj)
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

func (server *Server) Accept(socket_tag string, struct_type reflect.Type, function func(interface{})) {
	// create location in events map if needed?
	server.Events[socket_tag][struct_type] = append(server.Events[socket_tag][struct_type], function)
}

func (server *Server) AcceptRequest(socket_tag string, struct_type reflect.Type, function func(*net.Conn, uint16, interface{})) {
	// create location in requests map if needed?
	server.Requests[socket_tag][struct_type] = append(server.Events[socket_tag][struct_type], function)
}

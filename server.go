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
	Types		map[uint16]func()
	TypeCodes	map[reflect.Type]uint16
	Tag			func(*net.Conn)
	Events		map[string]map[reflect.Type][]func(interface{})
	Requests	map[string]map[reflect.Type][]func(*net.Conn, uint16, interface{})
}

type Respose struct {
	RequestID	uint16
	Type		uint16
	Data		string
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

func (socket *net.Conn) nextStruct(server Server) (interface{}, error) {
	header := make([]byte, 6)
	_, err := socket.Read(header)
	if err != nil { return err }
	
	type_bytes := header[:1]	// First two bytes are struct type
	size_bytes := header[2:]	// Next four bytes are struct size
	
	//type_int := binary.LittleEndian.ReadUint16(type_bytes)
	//size_int := binary.LittleEndian.ReadUint16(size_bytes)
	
	struct_data := make([]byte, size_int)
	_, err := socket.Read(struct_data)
	if err != nil { return err }
	
	recieved_struct := server.Types[type_int](struct_data)	// ensure location in map not nil
	
	return recieved_struct, nil	
}

func (server *Server) readStructs(socket *net.Conn) {
	obj, err := socket.nextStruct()
	if err != nil { return }
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

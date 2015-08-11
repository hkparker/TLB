package tlj

import (
	"fmt"
	"reflect"
	"encoding/json"
	"encoding/binary"
)

type Server struct {
	Listener		*net.UnixListener
	Tag				func(*net.Conn)
	Tags			map[*net.Conn][]string
	Types			map[uint16]func()
	TypeCodes		map[reflect.Type]uint16
	Events			map[string]map[reflect.Type][]func(interface{})
	Requests		map[string]map[reflect.Type][]func(*net.Conn, uint16, interface{})
	FailedServer	chan error
	FailedSockets	chan *net.Conn
}

func NewServer(listener *net.UnixListener, tag func(*net.Conn), types map[uint16]func(), type_codes map[reflect.Type]uint16) Server {
	server := Server {
		Listener:		listener,
		Tag:			tag,
		Tags:			make(map[*net.Conn][]string),
		Types:			types,
		TypeCodes:		type_codes,
		Events:			make(map[string]map[reflect.Type][]func(interface{})),
		Requests:		make(map[string]map[reflect.Type][]func(*net.Conn, uint16, interface{})),
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

func (server *Server) AcceptRequest(socket_tag string, struct_type reflect.Type, function func(*net.Conn, uint16, interface{})) {
	// create location in requests map if needed?
	server.Requests[socket_tag][struct_type] = append(server.Events[socket_tag][struct_type], function)
}

func (server *Server) Respond(socket, request_id, object) error {
	response_bytes, err := formatCapsule(object, request_id)
	if err != nil { return err }
	
	err = socket.Write(response_bytes)
	if err != nil {
		server.FailedSockets <- socket
		delete(server.Tags, socket)
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

func (socket *net.Conn) nextStruct(server Server) (interface{}, []string, error) {
	header := make([]byte, 6)
	_, err := socket.Read(header)
	if err != nil { return nil, nil, err }
	
	type_bytes := header[:2]
	size_bytes := header[2:]
	
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
		if err != nil {
			server.FailedSockets <- socket
			delete(server.Tags, socket)
			return
		}
		if obj == nil {
			continue
		} else if request.TypeOf(obj) == tlj.Capsule {	// or == request.TypeOf(Capsule{})
			for tag := range(tags) {
				//server.Requests[tag][reflect.TypeOf(obj)]  // make sure this isn't nil?
				for function := range(server.Requests[tag][reflect.TypeOf(obj)]) {
					request_id := obj.RequestID
					struct_type := obj.Type
					recieved_struct := server.Types[struct_type](obj.Data)
					if recieved_struct != nil { go function(socket, request_id, recieved_struct) }
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

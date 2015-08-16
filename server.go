package tlj

import (
	"net"
	"reflect"
)

type Server struct {
	Listener		*net.UnixListener
	TypeStore		*TypeStore
	Tag				func(*net.Conn)
	Tags			map[net.Conn][]string
	Events			map[string]map[uint16][]func(interface{})
	Requests		map[string]map[uint16][]func(interface{}, *Responder)
	FailedServer	chan error
	FailedSockets	chan net.Conn
}

func NewServer(listener *net.UnixListener, tag func(*net.Conn), type_store *TypeStore) Server {
	server := Server {
		Listener:		listener,
		TypeStore:		type_store,
		Tag:			tag,
		Tags:			make(map[net.Conn][]string),
		Events:			make(map[string]map[uint16][]func(interface{})),
		Requests:		make(map[string]map[uint16][]func(interface{}, *Responder)),
		FailedServer:	make(chan error, 1),
		FailedSockets:	make(chan net.Conn, 200),
	}
	go server.process()
	return server
}

func (server *Server) Accept(socket_tag string, struct_type reflect.Type, function func(interface{})) {
	// create location in events map if needed?  does tagging function know to create these locations?
	type_code, present := server.TypeStore.LookupCode(struct_type)
	if !present { return }
	server.Events[socket_tag][type_code] = append(server.Events[socket_tag][type_code], function)
}

func (server *Server) AcceptRequest(socket_tag string, struct_type reflect.Type, function func(interface{}, *Responder)) {
	// create location in requests map if needed?
	type_code, present := server.TypeStore.LookupCode(struct_type)
	if !present { return }
	server.Requests[socket_tag][type_code] = append(server.Requests[socket_tag][type_code], function)
}

type Responder struct {
	Server		*Server
	Socket		net.Conn
	RequestID	uint16
}

func (responder *Responder) Respond(object interface{}) error {
	response_bytes, err := formatCapsule(object, responder.Server.TypeStore, responder.RequestID)
	if err != nil { return err }
	
	_, err = responder.Socket.Write(response_bytes)
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
			break
		}
		server.Tags[socket] = make([]string, 0)
		server.Tag(&socket)
		go server.readStructs(socket)
	}
}

func (server *Server) readStructs(socket net.Conn) {
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
		} else if reflect.TypeOf(obj) == reflect.TypeOf(Capsule{}) {
			for tag := range(tags) {
				obj_value := reflect.Indirect(reflect.ValueOf(obj))
				embedded_request_id := uint16(obj_value.FieldByName("RequestID").Uint())
				embedded_type_code := uint16(obj_value.FieldByName("Type").Uint())
				embedded_data := obj_value.FieldByName("Data").String()
				if server.Requests[tag][embedded_type_code] == nil { continue }		// depends on how it was created
				for function := range(server.Requests[tag][embedded_type_code]) {
					responder := Responder {
						Server:		server,
						Socket:		socket,
						RequestID:	embedded_request_id,
					}
					recieved_struct := server.TypeStore.BuildType(embedded_type_code, []byte(embedded_data))		// base64 decode?
					if recieved_struct != nil { go function(recieved_struct, responder) }
				}
			}
		} else {
			for tag := range(tags) {
				recieved_type, present := server.TypeStore.LookupCode(reflect.TypeOf(obj))
				if !present { continue }
				if server.Events[tag][recieved_type] == nil { continue }			// depends on how it was created
				for function := range(server.Events[tag][recieved_type]) {
					go function(obj)
				}
			}
		}
	}
}

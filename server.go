package tlj

type Server struct {
	Sockets		map[string][]*net.Conn // socket interface
	Server		*net.UnixListener
	Events		map[reflect.Type][]func(interface{})
	Tag			func(*net.Conn)
	//Requests	map[string]map[reflect.Type][]func(interface{})  //  string tag -> map from types to funcs
}



func (server *Server) process(socket) {
}

func (server *Server) Accept(socket_tag string, struct_type reflect.Type, function func(interface{})) {
}

func (server *Server) AcceptRequest(socket_tag string, struct_type reflect.Type, function func(uint16, interface{})) {
}

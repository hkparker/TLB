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



// Accept(struct) vs AcceptRequest(struct, response chan?)

package tlj

type Server struct {
	Listener	*net.UnixListener
	Sockets		map[string][]*net.Conn							// sockets tagged with strings
	Tag			func(*net.Conn)
	Events		map[string]map[reflect.Type][]func(interface{})						//  string tag -> map from types to funcs
	Requessts	map[string]map[reflect.Type][]func(*net.Conn, uint16, interface{})	//  string tag -> map from types to funcs
}

func NewServer(listener *net.UnixListener, tag func(*net.Conn)) Server {
	
}

func (server *Server) process(socket) {
	// accept new connections from Listener
	
	// read from listener
}

func (server *Server) Accept(socket_tag string, struct_type reflect.Type, function func(interface{})) {
}

func (server *Server) AcceptRequest(socket_tag string, struct_type reflect.Type, function func(*net.Conn, uint16, interface{})) {
	// responses must go back the proper socket (also pass as a parameter?)
}

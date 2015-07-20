package tlj

type TLJServer struct {
	Sockets		map[string]tls.Conn // socket interface
	Server		*net.UnixListener
	Events		map[type][]func(type) //huh...
	Tag			func(socket)void
}



func (tljserver *TLJServer) Process(socket) {
}



// Accept(struct) vs AcceptRequest(struct, response chan?)

package tlj

import (
	"testing"
	"net"
	"reflect"
	"github.com/twinj/uuid"
	"os"
)

func TagSocketAll(socket *net.Conn, server *Server) {
    server.Tags[socket] = append(server.Tags[socket], "all")
    server.Sockets["all"] = append(server.Sockets["all"], socket)
}

func TestServerIsCorrectType(t *testing.T) {
	server_filename := "server_test-ipc-" + uuid.NewV4().String()
	listener, err := net.Listen("unix", server_filename)
	if err != nil {
		t.Errorf("could not start unix server")
	}
	defer listener.Close()
	defer os.RemoveAll(server_filename)
	type_store := NewTypeStore()
	server := NewServer(listener, TagSocketAll, &type_store)
	if reflect.TypeOf(server) != reflect.TypeOf(Server{}) {
		t.Errorf("return value of NewServer() != tlj.Server")
	} 
}

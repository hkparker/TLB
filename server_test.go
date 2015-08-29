package tlj

import (
	"testing"
	"net"
	"reflect"
	"github.com/twinj/uuid"
	"os"
	"time"
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

func TestServerCanReceiveAndTagConnection(t *testing.T) {
	server_filename := "server_test-ipc-" + uuid.NewV4().String()
	listener, err := net.Listen("unix", server_filename)
	if err != nil {
		t.Errorf("could not start unix server")
	}
	defer listener.Close()
	defer os.RemoveAll(server_filename)
	type_store := NewTypeStore()
	server := NewServer(listener, TagSocketAll, &type_store)
	client_socket, err := net.Dial("unix", server_filename)
	if err != nil {
		t.Errorf("could not connect to unix server")
	}
	defer client_socket.Close()
	time.Sleep(5 * time.Millisecond)		// wait for server to process incoming connection
	server_conns := server.Sockets["all"]
	if len(server_conns) != 1 {
		t.Errorf("socket did not get tagged as all")
	}
	if server.Tags[server_conns[0]][0] != "all" {
		t.Errorf("socket did not get tagged as all")
	}
}

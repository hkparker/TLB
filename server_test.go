package tlj_test

import (
	. "github.com/hkparker/TLJ"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
	"time"
)

func TagSocketAll(socket net.Conn, server *Server) {
    server.Tags[socket] = append(server.Tags[socket], "all")
    server.Sockets["all"] = append(server.Sockets["all"], socket)
}

var _ = Describe("Server", func() {

        var (
                type_store              TypeStore
                //populated_type_store    TypeStore
                //capsule                 Capsule
                //thingy                  Thingy
        )

        BeforeEach(func() {
                type_store = NewTypeStore()
                //populated_type_store = NewTypeStore()
                //inst_type := reflect.TypeOf(Thingy{})
                //ptr_type := reflect.TypeOf(&Thingy{})
                //populated_type_store.AddType(inst_type, ptr_type, BuildThingy)
                //capsule = Capsule {
                //        RequestID:      1,
                //        Type:           1,
                //        Data:           "test",
                //}
                //thingy = Thingy {
                //        Name:   "test",
                //        ID:     1,
                //}
        })


	It("can receive sockets and tag them", func() {
		listener, err := net.Listen("tcp", "localhost:5000")
		Expect(err).To(BeNil())
		defer listener.Close()
		server := NewServer(listener, TagSocketAll, &type_store)
		client_socket, err := net.Dial("tcp", "localhost:5000")
		Expect(err).To(BeNil())
		defer client_socket.Close()
		time.Sleep(5 * time.Millisecond)	// give the server time to process incoming connection
		server_conns := server.Sockets["all"]
		Expect(len(server_conns)).To(Equal(1))
		Expect(server.Tags[server_conns[0]][0]).To(Equal("all"))
	})

	It("can insert a client socket into the server", func() {
		listener, err := net.Listen("tcp", "localhost:5001")
		Expect(err).To(BeNil())
		go func() {
			listener.Accept()
		}()
		defer listener.Close()
		server := NewServer(listener, TagSocketAll, &type_store)
		other_listener, err := net.Listen("tcp", "localhost:5002")
		Expect(err).To(BeNil())
		defer other_listener.Close()
		client_socket, err := net.Dial("tcp", "localhost:5001")
		defer client_socket.Close()
		server.Insert(client_socket)
		server_conns := server.Sockets["all"]
		Expect(len(server_conns)).To(Equal(1))
		Expect(server.Tags[server_conns[0]][0]).To(Equal("all"))
	})

	It("", func() {

	})

	It("", func() {

	})

	It("", func() {

	})
})

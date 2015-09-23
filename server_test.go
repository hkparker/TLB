package tlj_test

import (
	. "github.com/hkparker/TLJ"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
	"time"
	"reflect"
	"os"
	"fmt"
)

func TagSocketAll(socket net.Conn, server *Server) {
    server.Tags[socket] = append(server.Tags[socket], "all")
    server.Sockets["all"] = append(server.Sockets["all"], socket)
}

var _ = Describe("Server", func() {

        var (
                type_store              TypeStore
		populated_type_store	TypeStore
                thingy                  Thingy
        )

        BeforeEach(func() {
                type_store = NewTypeStore()
                populated_type_store = NewTypeStore()
                inst_type := reflect.TypeOf(Thingy{})
                ptr_type := reflect.TypeOf(&Thingy{})
                populated_type_store.AddType(inst_type, ptr_type, BuildThingy)
                thingy = Thingy {
                        Name:   "test",
                        ID:     1,
                }
        })


	It("can receive sockets and tag them", func() {
		listener, err := net.Listen("tcp", "localhost:5002")
		Expect(err).To(BeNil())
		defer listener.Close()
		server := NewServer(listener, TagSocketAll, &type_store)
		client_socket, err := net.Dial("tcp", "localhost:5002")
		Expect(err).To(BeNil())
		defer client_socket.Close()
		server_conns := server.Sockets["all"]
		Expect(len(server_conns)).To(Equal(1))
		Expect(server.Tags[server_conns[0]][0]).To(Equal("all"))
	})

	It("can insert a client socket into the server", func() {
		listener, err := net.Listen("tcp", "localhost:5003")
		Expect(err).To(BeNil())
		defer listener.Close()
		server := NewServer(listener, TagSocketAll, &type_store)
		other_listener, err := net.Listen("tcp", "localhost:5004")
		Expect(err).To(BeNil())
		defer other_listener.Close()
		client_socket, err := net.Dial("tcp", "localhost:5004")
		defer client_socket.Close()
		server.Insert(client_socket)
		server_conns := server.Sockets["all"]
		Expect(len(server_conns)).To(Equal(1))
		Expect(server.Tags[server_conns[0]][0]).To(Equal("all"))
	})

	It("can run the accept callbacks", func() {
		listener, err := net.Listen("tcp", "localhost:5005")
		Expect(err).To(BeNil())
		defer listener.Close()
		server := NewServer(listener, TagSocketAll, &populated_type_store)
		server.Accept("all", reflect.TypeOf(Thingy{}), func(iface interface{}) {	
			if received_thingy, correct_type :=  iface.(*Thingy); correct_type {
				Expect(received_thingy.Name).To(Equal(thingy.Name))
				Expect(received_thingy.ID).To(Equal(thingy.ID))
				f, err := os.Create(fmt.Sprintf("accept-%s-%d.0", thingy.Name, thingy.ID))
				Expect(err).To(BeNil())
				f.Close()
			} else {
				Expect(correct_type).To(Equal(true))
			}
		})
		server.Accept("all", reflect.TypeOf(Thingy{}), func(iface interface{}) {
			if received_thingy, correct_type :=  iface.(*Thingy); correct_type {
				Expect(received_thingy.Name).To(Equal(thingy.Name))
				Expect(received_thingy.ID).To(Equal(thingy.ID))
				f, err := os.Create(fmt.Sprintf("accept-%s-%d.1", thingy.Name, thingy.ID))
				Expect(err).To(BeNil())
				f.Close()
			} else {
				Expect(correct_type).To(Equal(true))
			}
		})
		Expect(server.Events["all"]).ToNot(BeNil())
		Expect(server.Events["all"][1]).ToNot(BeNil())
		Expect(len(server.Events["all"][1])).To(Equal(2))
		client_socket, err := net.Dial("tcp", "localhost:5005")
		Expect(err).To(BeNil())
		defer client_socket.Close()
		thingy_bytes, err := Format(thingy, &populated_type_store)
		Expect(err).To(BeNil())
		client_socket.Write(thingy_bytes)
		time.Sleep(10 * time.Millisecond)	// give the server time to run the callbacks
		client_socket2, err := net.Dial("tcp", "localhost:5005")
		Expect(err).To(BeNil())
		defer client_socket2.Close()
		_, err = os.Stat("accept-test-1.0")
		Expect(err).To(BeNil())
		_, err = os.Stat("accept-test-1.1")
		Expect(err).To(BeNil())
		os.RemoveAll("accept-test-1.0")
		os.RemoveAll("accept-test-1.1")
	})

	It("can run accept request callbacks", func() {
		
	})

	Describe("Responder", func() {

		It("can be used to send a response", func() {

		})
	})
})

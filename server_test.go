package tlj_test

import (
	"encoding/json"
	"fmt"
	. "github.com/hkparker/TLJ"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
	"reflect"
	"sync"
)

func TagSocketAll(socket net.Conn, server *Server) {
	server.TagSocket(socket, "all")
}

var _ = Describe("Server", func() {

	var (
		type_store           TypeStore
		populated_type_store TypeStore
		thingy               Thingy
	)

	BeforeEach(func() {
		type_store = NewTypeStore()
		populated_type_store = NewTypeStore()
		inst_type := reflect.TypeOf(Thingy{})
		ptr_type := reflect.TypeOf(&Thingy{})
		populated_type_store.AddType(inst_type, ptr_type, BuildThingy)
		thingy = Thingy{
			Name: "test",
			ID:   1,
		}
	})

	Describe("Accept", func() {
		It("can run the accept callbacks", func() {
			listener, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer listener.Close()
			server := NewServer(listener, TagSocketAll, populated_type_store)
			first_chan := make(chan string)
			second_chan := make(chan string)
			server.Accept("all", reflect.TypeOf(Thingy{}), func(iface interface{}, _ TLJContext) {
				if received_thingy, correct_type := iface.(*Thingy); correct_type {
					first_chan <- fmt.Sprintf("accept-%s-%d.0", received_thingy.Name, received_thingy.ID)
				} else {
					Expect(correct_type).To(Equal(true))
				}
			})
			server.Accept("all", reflect.TypeOf(Thingy{}), func(iface interface{}, _ TLJContext) {
				if received_thingy, correct_type := iface.(*Thingy); correct_type {
					second_chan <- fmt.Sprintf("accept-%s-%d.1", received_thingy.Name, received_thingy.ID)
				} else {
					Expect(correct_type).To(Equal(true))
				}
			})
			Expect(server.Events["all"]).ToNot(BeNil())
			Expect(server.Events["all"][1]).ToNot(BeNil())
			Expect(len(server.Events["all"][1])).To(Equal(2))
			client_socket, err := net.Dial("tcp", listener.Addr().String())
			Expect(err).To(BeNil())
			defer client_socket.Close()
			thingy_bytes, err := populated_type_store.Format(thingy)
			Expect(err).To(BeNil())
			client_socket.Write(thingy_bytes)
			client_socket2, err := net.Dial("tcp", listener.Addr().String())
			Expect(err).To(BeNil())
			defer client_socket2.Close()
			first_incoming_thingy := <-first_chan
			second_incoming_thingy := <-second_chan
			Expect(first_incoming_thingy).To(Equal("accept-test-1.0"))
			Expect(second_incoming_thingy).To(Equal("accept-test-1.1"))
		})
	})

	Describe("AcceptRequest", func() {
		It("can run accept request callbacks", func() {
			listener, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer listener.Close()
			server := NewServer(listener, TagSocketAll, populated_type_store)
			first_chan := make(chan string)
			second_chan := make(chan string)
			server.AcceptRequest("all", reflect.TypeOf(Thingy{}), func(iface interface{}, _ TLJContext) {
				if received_thingy, correct_type := iface.(*Thingy); correct_type {
					first_chan <- fmt.Sprintf("acceptrequest-%s-%d.0", received_thingy.Name, received_thingy.ID)
				} else {
					Expect(correct_type).To(Equal(true))
				}
			})
			server.AcceptRequest("all", reflect.TypeOf(Thingy{}), func(iface interface{}, _ TLJContext) {
				if received_thingy, correct_type := iface.(*Thingy); correct_type {
					second_chan <- fmt.Sprintf("acceptrequest-%s-%d.1", received_thingy.Name, received_thingy.ID)
				} else {
					Expect(correct_type).To(Equal(true))
				}
			})
			Expect(server.Requests["all"]).ToNot(BeNil())
			Expect(server.Requests["all"][1]).ToNot(BeNil())
			Expect(len(server.Requests["all"][1])).To(Equal(2))
			client_socket, err := net.Dial("tcp", listener.Addr().String())
			Expect(err).To(BeNil())
			defer client_socket.Close()
			capsule_bytes, err := populated_type_store.FormatCapsule(thingy, 1)
			Expect(err).To(BeNil())
			client_socket.Write(capsule_bytes)
			client_socket2, err := net.Dial("tcp", listener.Addr().String())
			Expect(err).To(BeNil())
			defer client_socket2.Close()
			first_incoming_thingy := <-first_chan
			second_incoming_thingy := <-second_chan
			Expect(first_incoming_thingy).To(Equal("acceptrequest-test-1.0"))
			Expect(second_incoming_thingy).To(Equal("acceptrequest-test-1.1"))
		})
	})

	Describe("ExcludeString", func() {
		It("excludes strings by value", func() {
			set1 := []string{"a", "b", "c"}
			set2 := ExcludeString(set1, "b")
			Expect(set2).To(Equal([]string{"a", "c"}))
		})
	})

	Describe("UntagSocket", func() {
		It("Untags a socket", func() {
			listener, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer listener.Close()
			server := NewServer(listener, TagSocketAll, type_store)
			other_listener, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer other_listener.Close()
			client_socket, err := net.Dial("tcp", other_listener.Addr().String())
			defer client_socket.Close()
			server.Insert(client_socket)
			server_conns := server.Sockets["all"]
			Expect(len(server_conns)).To(Equal(1))
			Expect(server.Tags[server_conns[0]][0]).To(Equal("all"))
			server.TagSocket(client_socket, "foo")
			Expect(len(server.Sockets["foo"])).To(Equal(1))
			server.UntagSocket(client_socket, "foo")
			Expect(len(server.Sockets["foo"])).To(Equal(0))
		})
	})

	Describe("Insert", func() {
		It("can insert a client socket into the server", func() {
			listener, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer listener.Close()
			server := NewServer(listener, TagSocketAll, type_store)
			other_listener, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer other_listener.Close()
			client_socket, err := net.Dial("tcp", other_listener.Addr().String())
			defer client_socket.Close()
			server.Insert(client_socket)
			server_conns := server.Sockets["all"]
			Expect(len(server_conns)).To(Equal(1))
			Expect(server.Tags[server_conns[0]][0]).To(Equal("all"))
		})
	})

	Describe("Delete", func() {
		It("can delet a socket from all tags", func() {
			listener, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer listener.Close()
			server := NewServer(listener, TagSocketAll, type_store)
			other_listener, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer other_listener.Close()
			client_socket, err := net.Dial("tcp", other_listener.Addr().String())
			defer client_socket.Close()
			server.Insert(client_socket)
			server_conns := server.Sockets["all"]
			Expect(len(server_conns)).To(Equal(1))
			Expect(server.Tags[server_conns[0]][0]).To(Equal("all"))
			server.TagSocket(client_socket, "foo")
			Expect(len(server.Sockets["foo"])).To(Equal(1))
			server.Delete(client_socket)
			Expect(len(server.Sockets["foo"])).To(Equal(0))
			Expect(server.Tags[client_socket]).To(BeNil())
		})
	})

	Describe("TLJContext", func() {
		It("can be used to send a response", func() {
			listener, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer listener.Close()
			sockets := make(chan net.Conn, 1)
			go func() {
				conn, _ := listener.Accept()
				sockets <- conn
			}()
			client, err := net.Dial("tcp", listener.Addr().String())
			Expect(err).To(BeNil())
			defer client.Close()
			server_side := <-sockets
			server := NewServer(listener, TagSocketAll, populated_type_store)
			responder := Responder{
				RequestID: 1,
				WriteLock: sync.Mutex{},
			}
			context := TLJContext{
				Server:    &server,
				Socket:    server_side,
				Responder: responder,
			}
			err = context.Respond(thingy)
			Expect(err).To(BeNil())
			iface, err := populated_type_store.NextStruct(client, TLJContext{})
			Expect(err).To(BeNil())
			if response, correct_type := iface.(*Capsule); correct_type {
				Expect(response.RequestID).To(Equal(uint16(1)))
				Expect(response.Type).To(Equal(uint16(1)))
				restored_thing := &Thingy{}
				err = json.Unmarshal([]byte(response.Data), &restored_thing)
				Expect(err).To(BeNil())
				Expect(restored_thing.ID).To(Equal(thingy.ID))
				Expect(restored_thing.Name).To(Equal(thingy.Name))
			} else {
				Expect(correct_type).To(Equal(true))
			}
		})
	})
})

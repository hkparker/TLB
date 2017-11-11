package tlb_test

import (
	"encoding/binary"
	. "github.com/hkparker/TLB"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/mgo.v2/bson"
	"net"
	"reflect"
	"time"
)

var _ = Describe("Client", func() {

	var (
		populated_type_store TypeStore
		thingy               Thingy
	)

	BeforeEach(func() {
		populated_type_store = NewTypeStore()
		inst_type := reflect.TypeOf(Thingy{})
		ptr_type := reflect.TypeOf(&Thingy{})
		populated_type_store.AddType(inst_type, ptr_type, BuildThingy)
		thingy = Thingy{
			Name: "test",
			ID:   1,
		}
	})

	Describe("Message", func() {
		It("can use message to write a struct to a socket", func() {
			listener, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer listener.Close()
			sockets := make(chan net.Conn, 1)
			go func() {
				conn, _ := listener.Accept()
				sockets <- conn
			}()
			client_side, err := net.Dial("tcp", listener.Addr().String())
			Expect(err).To(BeNil())
			defer client_side.Close()
			server_side := <-sockets
			client := NewClient(client_side, populated_type_store, true)
			err = client.Message(thingy)
			Expect(err).To(BeNil())
			iface, err := populated_type_store.NextStruct(server_side, TLBContext{})
			Expect(err).To(BeNil())
			if received_thingy, correct_type := iface.(*Thingy); correct_type {
				Expect(received_thingy.ID).To(Equal(thingy.ID))
				Expect(received_thingy.Name).To(Equal(thingy.Name))
			} else {
				Expect(correct_type).To(Equal(true))
			}
		})
	})

	Describe("Request", func() {
		It("can use request to write a capsule to a socket", func() {
			listener, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer listener.Close()
			sockets := make(chan net.Conn, 1)
			go func() {
				conn, _ := listener.Accept()
				sockets <- conn
			}()
			client_side, err := net.Dial("tcp", listener.Addr().String())
			Expect(err).To(BeNil())
			defer client_side.Close()
			server_side := <-sockets
			client := NewClient(client_side, populated_type_store, true)
			_, err = client.Request(thingy)
			Expect(err).To(BeNil())
			iface, err := populated_type_store.NextStruct(server_side, TLBContext{})
			if capsule, correct_type := iface.(*Capsule); correct_type {
				Expect(capsule.Type).To(Equal(uint16(1)))
				restored_thing := &Thingy{}
				err = bson.Unmarshal([]byte(capsule.Data), &restored_thing)
				Expect(err).To(BeNil())
				Expect(restored_thing.ID).To(Equal(thingy.ID))
				Expect(restored_thing.Name).To(Equal(thingy.Name))
			} else {
				Expect(correct_type).To(Equal(true))
			}
		})
	})

	Describe("StreamWriter", func() {
		Describe("Write", func() {
			It("outputs the correct format", func() {
				listener, err := net.Listen("tcp", "localhost:0")
				Expect(err).To(BeNil())
				defer listener.Close()
				sockets := make(chan net.Conn, 1)
				go func() {
					conn, err := listener.Accept()
					Expect(err).To(BeNil())
					sockets <- conn
				}()
				conn, err := net.Dial("tcp", listener.Addr().String())
				Expect(err).To(BeNil())
				client := <-sockets
				stream_writer, err := NewStreamWriter(
					conn,
					populated_type_store,
					reflect.TypeOf(Thingy{}),
				)
				Expect(err).To(BeNil())
				err = stream_writer.Write(thingy)
				Expect(err).To(BeNil())
				thingy_bytes := make([]byte, 50)
				n, err := client.Read(thingy_bytes)
				Expect(err).To(BeNil())
				read_bytes := thingy_bytes[:n]
				type_bytes := read_bytes[:2]
				size_bytes := read_bytes[2:6]
				thingy_data := read_bytes[6:]
				type_int := binary.LittleEndian.Uint16(type_bytes)
				size_int := binary.LittleEndian.Uint32(size_bytes)
				Expect(type_int).To(Equal(uint16(1)))
				Expect(size_int).To(Equal(uint32(len(thingy_data))))
				restored_thingy := Thingy{}
				err = bson.Unmarshal(thingy_data, &restored_thingy)
				Expect(err).To(BeNil())
				Expect(restored_thingy.ID).To(Equal(1))
				Expect(restored_thingy.Name).To(Equal("test"))
			})
		})
	})

	Describe("Request", func() {
		Describe("OnResponse", func() {
			It("can run on response callbacks", func() {
				listener, err := net.Listen("tcp", "localhost:0")
				Expect(err).To(BeNil())
				defer listener.Close()
				server := NewServer(listener, TagSocketAll, populated_type_store)
				server.AcceptRequest("all", reflect.TypeOf(Thingy{}), func(iface interface{}, context TLBContext) {
					resp := Thingy{
						ID:   2,
						Name: "response!",
					}
					time.Sleep(1 * time.Second)
					context.Respond(resp)
				})
				client_socket, err := net.Dial("tcp", listener.Addr().String())
				Expect(err).To(BeNil())
				defer client_socket.Close()
				client := NewClient(client_socket, populated_type_store, false)
				request, err := client.Request(thingy)
				Expect(err).To(BeNil())
				run_chan := make(chan string)
				request.OnResponse(reflect.TypeOf(&Thingy{}), func(iface interface{}) {
					if response, correct_type := iface.(*Thingy); correct_type {
						Expect(response.ID).To(Equal(2))
						Expect(response.Name).To(Equal("response!"))
						run_chan <- "ran"
					} else {
						Expect(correct_type).To(Equal(true))
					}
				})
				ran := <-run_chan
				Expect(ran).To(Equal("ran"))
			})
		})
	})
})

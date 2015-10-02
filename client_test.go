package tlj_test

import (
	. "github.com/hkparker/TLJ"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"reflect"
	"net"
	"encoding/json"
	"time"
)

var _ = Describe("Client", func() {

	var (
		populated_type_store    TypeStore
		thingy                  Thingy
	)

	BeforeEach(func() {
		populated_type_store = NewTypeStore()
		inst_type := reflect.TypeOf(Thingy{})
		ptr_type := reflect.TypeOf(&Thingy{})
		populated_type_store.AddType(inst_type, ptr_type, BuildThingy)
		thingy = Thingy {
			Name:   "test",
			ID:     1,
		}
	})

	It("can use message to write a struct to a socket", func() {
		listener, err := net.Listen("tcp", "localhost:5008")
		Expect(err).To(BeNil())
		defer listener.Close()
		sockets := make(chan net.Conn, 1)
		go func() {
			conn, _ := listener.Accept()
			sockets <- conn
		}()
		client_side, err := net.Dial("tcp", "localhost:5008")
		Expect(err).To(BeNil())
		defer client_side.Close()
		server_side := <- sockets
		client := NewClient(client_side, &populated_type_store)
		err = client.Message(thingy)
		Expect(err).To(BeNil())
		iface, err := NextStruct(server_side, &populated_type_store)
		if received_thingy, correct_type := iface.(*Thingy); correct_type {
			Expect(received_thingy.ID).To(Equal(thingy.ID))
			Expect(received_thingy.Name).To(Equal(thingy.Name))
		} else {
			Expect(correct_type).To(Equal(true))
		}
	})

	It("can use request to write a capsule to a socket", func() {	
		listener, err := net.Listen("tcp", "localhost:5009")
		Expect(err).To(BeNil())
		defer listener.Close()
		sockets := make(chan net.Conn, 1)
		go func() {
			conn, _ := listener.Accept()
			sockets <- conn
		}()
		client_side, err := net.Dial("tcp", "localhost:5009")
		Expect(err).To(BeNil())
		defer client_side.Close()
		server_side := <- sockets
		client := NewClient(client_side, &populated_type_store)
		_, err = client.Request(thingy)
		Expect(err).To(BeNil())
		iface, err := NextStruct(server_side, &populated_type_store)
		if capsule, correct_type := iface.(*Capsule); correct_type {
			Expect(capsule.Type).To(Equal(uint16(1)))
			restored_thing := &Thingy{}
			err = json.Unmarshal([]byte(capsule.Data), &restored_thing)
			Expect(err).To(BeNil())
			Expect(restored_thing.ID).To(Equal(thingy.ID))
			Expect(restored_thing.Name).To(Equal(thingy.Name))
		} else {
			Expect(correct_type).To(Equal(true))
		}
	})

	It("can run on response callbacks", func() {	
		listener, err := net.Listen("tcp", "localhost:5010")
		Expect(err).To(BeNil())
		defer listener.Close()
		server := NewServer(listener, TagSocketAll, &populated_type_store)
		server.AcceptRequest("all", reflect.TypeOf(Thingy{}), func(iface interface{}, responder Responder) {
			resp := Thingy {
				ID:	2,
				Name:	"response!",
			}
			time.Sleep(1 * time.Second)
			responder.Respond(resp)
		})
		client_socket, err := net.Dial("tcp", "localhost:5010")
		Expect(err).To(BeNil())
		defer client_socket.Close()
		client := NewClient(client_socket, &populated_type_store)
		request, err := client.Request(thingy)
		Expect(err).To(BeNil())
		run_chan := make(chan string)
		request.OnResponse(reflect.TypeOf(&Thingy{}), func(iface interface{}){
			if response, correct_type := iface.(*Thingy); correct_type {
				Expect(response.ID).To(Equal(2))
				Expect(response.Name).To(Equal("response!"))
				run_chan <- "ran"
			} else {
				Expect(correct_type).To(Equal(true))
			}
		})
		ran := <- run_chan
		Expect(ran).To(Equal("ran"))
	})
})

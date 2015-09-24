package tlj_test

import (
	. "github.com/hkparker/TLJ"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"reflect"
	"net"
	
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
	
	})

	It("can run on response callbacks", func() {
	
	})
})

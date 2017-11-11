package tlb_test

import (
	"encoding/binary"
	. "github.com/hkparker/TLB"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/mgo.v2/bson"
	"net"
	"reflect"
)

type Thingy struct {
	Name string
	ID   int
}

func BuildThingy(data []byte, _ TLBContext) interface{} {
	thing := &Thingy{}
	err := bson.Unmarshal(data, &thing)
	if err != nil {
		return nil
	}
	return thing
}

var _ = Describe("TypeStore", func() {

	var (
		type_store           TypeStore
		populated_type_store TypeStore
		capsule              Capsule
		thingy               Thingy
	)

	BeforeEach(func() {
		type_store = NewTypeStore()
		populated_type_store = NewTypeStore()
		inst_type := reflect.TypeOf(Thingy{})
		ptr_type := reflect.TypeOf(&Thingy{})
		populated_type_store.AddType(inst_type, ptr_type, BuildThingy)
		capsule = Capsule{
			RequestID: 1,
			Type:      1,
			Data:      "test",
		}
		thingy = Thingy{
			Name: "test",
			ID:   1,
		}
	})

	Describe("NewTypeStore", func() {
		It("can unmarshal a capsule when created", func() {
			capsule_bytes, err := bson.Marshal(capsule)
			Expect(err).To(BeNil())
			iface := type_store.BuildType(0, capsule_bytes, TLBContext{})
			if restored, correct_type := iface.(*Capsule); correct_type {
				Expect(restored.RequestID).To(Equal(capsule.RequestID))
				Expect(restored.Type).To(Equal(capsule.Type))
				Expect(restored.Data).To(Equal(capsule.Data))
			} else {
				Expect(correct_type).To(Equal(true))
			}
		})
	})

	Describe("AddType", func() {
		It("can add a new type", func() {
			inst_type := reflect.TypeOf(Thingy{})
			ptr_type := reflect.TypeOf(&Thingy{})
			err := type_store.AddType(inst_type, ptr_type, BuildThingy)
			Expect(err).To(BeNil())
			Expect(type_store.TypeCodes[inst_type]).To(Equal(uint16(1)))
			Expect(type_store.TypeCodes[ptr_type]).To(Equal(uint16(1)))
		})

		It("reports an error with nil instance type", func() {
			ptr_type := reflect.TypeOf(&Thingy{})
			err := type_store.AddType(nil, ptr_type, BuildThingy)
			Expect(err).ToNot(BeNil())
		})

		It("reports an error with nil pointer type", func() {
			inst_type := reflect.TypeOf(Thingy{})
			err := type_store.AddType(inst_type, nil, BuildThingy)
			Expect(err).ToNot(BeNil())
		})

		It("reports an error with nil builder", func() {
			inst_type := reflect.TypeOf(Thingy{})
			ptr_type := reflect.TypeOf(&Thingy{})
			err := type_store.AddType(inst_type, ptr_type, nil)
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("LookupCode", func() {
		It("correctly looks up codes", func() {
			code, present := type_store.LookupCode(reflect.TypeOf(Capsule{}))
			Expect(present).To(Equal(true))
			Expect(code).To(Equal(uint16(0)))
		})

		It("correctly reports incorrect codes", func() {
			code, present := type_store.LookupCode(reflect.TypeOf(Thingy{}))
			Expect(present).To(Equal(false))
			Expect(code).To(Equal(uint16(0)))
		})
	})

	Describe("BuildType", func() {
		It("can build a type", func() {
			thingy_bytes, err := bson.Marshal(thingy)
			Expect(err).To(BeNil())
			iface := populated_type_store.BuildType(1, thingy_bytes, TLBContext{})
			if restored, correct_type := iface.(*Thingy); correct_type {
				Expect(restored.Name).To(Equal(thingy.Name))
				Expect(restored.ID).To(Equal(thingy.ID))
			} else {
				Expect(correct_type).To(Equal(true))
			}
		})

		It("wont build bad codes", func() {
			iface := type_store.BuildType(1, make([]byte, 0), TLBContext{})
			Expect(iface).To(BeNil())
		})

		It("wont build unformatted data", func() {
			iface := type_store.BuildType(0, []byte("notbson"), TLBContext{})
			Expect(iface).To(BeNil())
		})
	})

	Describe("Format", func() {
		It("formats data correctly", func() {
			thingy_bytes, err := populated_type_store.Format(thingy)
			Expect(err).To(BeNil())
			type_bytes := thingy_bytes[:2]
			size_bytes := thingy_bytes[2:6]
			bson_data := thingy_bytes[6:]
			type_int := binary.LittleEndian.Uint16(type_bytes)
			size_int := binary.LittleEndian.Uint32(size_bytes)
			Expect(type_int).To(Equal(uint16(1)))
			Expect(size_int).To(Equal(uint32(len(bson_data))))

			actual_bytes, err := bson.Marshal(thingy)
			Expect(err).To(BeNil())
			Expect(bson_data).To(Equal(actual_bytes))

			restored_thing := &Thingy{}
			err = bson.Unmarshal(bson_data, &restored_thing)
			Expect(err).To(BeNil())
		})

		It("returns an error when type missing from store", func() {
			_, err := type_store.Format(thingy)
			Expect(err).ToNot(BeNil())
		})

		It("returns an error when the instance is nil", func() {
			_, err := type_store.Format(nil)
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("FormatCapsule", func() {
		It("formats capsules correctly", func() {
			capsule_bytes, err := populated_type_store.FormatCapsule(thingy, 1)
			Expect(err).To(BeNil())
			type_bytes := capsule_bytes[:2]
			size_bytes := capsule_bytes[2:6]
			capsule_data := capsule_bytes[6:]
			type_int := binary.LittleEndian.Uint16(type_bytes)
			size_int := binary.LittleEndian.Uint32(size_bytes)
			Expect(type_int).To(Equal(uint16(0)))
			Expect(size_int).To(Equal(uint32(len(capsule_data))))
			restored_capsule := &Capsule{}
			err = bson.Unmarshal(capsule_data, &restored_capsule)
			Expect(err).To(BeNil())
			Expect(restored_capsule.RequestID).To(Equal(uint16(1)))
			Expect(restored_capsule.Type).To(Equal(uint16(1)))
			restored_thing := &Thingy{}
			err = bson.Unmarshal([]byte(restored_capsule.Data), &restored_thing)
			Expect(err).To(BeNil())
			Expect(restored_thing.Name).To(Equal("test"))
			Expect(restored_thing.ID).To(Equal(1))
		})

		It("returns an error when type missing from type store", func() {
			_, err := type_store.FormatCapsule(thingy, 1)
			Expect(err).ToNot(BeNil())
		})

		It("returns an error when the instance is nil", func() {
			_, err := type_store.FormatCapsule(nil, 1)
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("NextStruct", func() {
		It("can read multiple structs", func() {
			sockets := make(chan net.Conn, 1)
			server, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer server.Close()
			go func() {
				conn, _ := server.Accept()
				sockets <- conn
			}()
			client, err := net.Dial("tcp", server.Addr().String())
			Expect(err).To(BeNil())
			defer client.Close()
			server_side := <-sockets
			unicode_thingy := Thingy{
				Name: "😃",
				ID:   2,
			}
			thingy_bytes, _ := populated_type_store.Format(thingy)
			unicode_thingy_bytes, _ := populated_type_store.Format(unicode_thingy)
			server_side.Write(thingy_bytes)
			server_side.Write(unicode_thingy_bytes)
			iface, err := populated_type_store.NextStruct(client, TLBContext{})
			Expect(err).To(BeNil())
			if restored_thingy, correct_type := iface.(*Thingy); correct_type {
				Expect(restored_thingy.Name).To(Equal(thingy.Name))
				Expect(restored_thingy.ID).To(Equal(thingy.ID))
			} else {
				Expect(correct_type).To(Equal(true))
			}
			iface, err = populated_type_store.NextStruct(client, TLBContext{})
			Expect(err).To(BeNil())
			if restored_thingy, correct_type := iface.(*Thingy); correct_type {
				Expect(restored_thingy.Name).To(Equal(unicode_thingy.Name))
				Expect(restored_thingy.ID).To(Equal(unicode_thingy.ID))
			} else {
				Expect(correct_type).To(Equal(true))
			}
		})

		It("reports an error when the socket is broken", func() {
			sockets := make(chan net.Conn, 1)
			server, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer server.Close()
			go func() {
				conn, _ := server.Accept()
				sockets <- conn
			}()
			client, err := net.Dial("tcp", server.Addr().String())
			Expect(err).To(BeNil())
			client.Close()
			_, err = type_store.NextStruct(client, TLBContext{})
			Expect(err).ToNot(BeNil())
		})

		It("returns nil and an error when the struct is missing from the type store", func() {
			sockets := make(chan net.Conn, 1)
			server, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer server.Close()
			go func() {
				conn, _ := server.Accept()
				sockets <- conn
			}()
			client, err := net.Dial("tcp", server.Addr().String())
			Expect(err).To(BeNil())
			defer client.Close()
			server_side := <-sockets
			thingy_bytes, _ := populated_type_store.Format(thingy)
			server_side.Write(thingy_bytes)
			iface, err := type_store.NextStruct(client, TLBContext{})
			Expect(iface).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("type code on received struct not in type store"))
		})

		It("returns an error when too few bytes are written", func() {
			sockets := make(chan net.Conn, 1)
			server, err := net.Listen("tcp", "localhost:0")
			Expect(err).To(BeNil())
			defer server.Close()
			go func() {
				conn, _ := server.Accept()
				sockets <- conn
			}()
			client, err := net.Dial("tcp", server.Addr().String())
			Expect(err).To(BeNil())
			defer client.Close()
			server_side := <-sockets
			server_side.Write([]byte{0x00, 0x01, 0x02})
			_, err = type_store.NextStruct(client, TLBContext{})
			Expect(err).ToNot(BeNil())
		})
	})
})

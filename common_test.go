package tlj_test

import (
	. "github.com/hkparker/TLJ"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
        "encoding/json"
	"encoding/binary"
	"reflect"
	"net"
)

type Thingy struct {
        Name    string
        ID              int
}

func BuildThingy(data []byte) interface{} {
        thing := &Thingy{}
        err := json.Unmarshal(data, &thing)
        if err != nil { return nil }
        return thing
}

var _ = Describe("Common", func() {

	var (
		type_store		TypeStore
		populated_type_store	TypeStore
		capsule			Capsule
		thingy			Thingy
	)

	BeforeEach(func() {
       	        type_store = NewTypeStore()
		populated_type_store = NewTypeStore()
		inst_type := reflect.TypeOf(Thingy{})
 		ptr_type := reflect.TypeOf(&Thingy{})
 		populated_type_store.AddType(inst_type, ptr_type, BuildThingy)
		capsule = Capsule {
			RequestID:	1,
			Type:		1,
			Data:		"test",
		}
		thingy = Thingy {
			Name:	"test",
			ID:	1,
		}
	})

	Describe("TypeStore", func() {

		It("should contain a functional capsule builder when created", func() {
			capsule_bytes, _ := json.Marshal(capsule)
			iface := type_store.BuildType(0, capsule_bytes)
			if restored, correct_type := iface.(*Capsule); correct_type {
				Expect(restored.RequestID).To(Equal(capsule.RequestID))
				Expect(restored.Type).To(Equal(capsule.Type))
				Expect(restored.Data).To(Equal(capsule.Data))
			} else {
				Expect(correct_type).To(Equal(true))
			}
		})

		It("can add a new type", func() {
			inst_type := reflect.TypeOf(Thingy{})
			ptr_type := reflect.TypeOf(&Thingy{})
			type_store.AddType(inst_type, ptr_type, BuildThingy)
			Expect(type_store.TypeCodes[inst_type]).To(Equal(uint16(1)))
			Expect(type_store.TypeCodes[ptr_type]).To(Equal(uint16(1)))
		})

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

		It("can build a type", func() {	
			thingy_bytes, _ := json.Marshal(thingy)
			iface := populated_type_store.BuildType(1, thingy_bytes)
			if restored, correct_type := iface.(*Thingy); correct_type {
				Expect(restored.Name).To(Equal(thingy.Name))
				Expect(restored.ID).To(Equal(thingy.ID))
			} else {
				Expect(correct_type).To(Equal(true))
			}
		})

		It("wont build bad codes", func() {
			iface := type_store.BuildType(1, make([]byte, 0))
			Expect(iface).To(BeNil())
		})

		It("wont build unformatted data", func() {
			iface := type_store.BuildType(0, []byte("notjson"))
			Expect(iface).To(BeNil())
		})
	})

	Describe("Format", func() {

		It("formats data correctly", func() {
			thingy_bytes, err := Format(thingy, &populated_type_store)
			Expect(err).To(BeNil())
			type_bytes := thingy_bytes[:2]
			size_bytes := thingy_bytes[2:6]
			json_data := thingy_bytes[6:]
			type_int := binary.LittleEndian.Uint16(type_bytes)
			size_int := binary.LittleEndian.Uint32(size_bytes)
			Expect(type_int).To(Equal(uint16(1)))
			Expect(size_int).To(Equal(uint32(len(json_data))))
			restored_thing := &Thingy{}
			err = json.Unmarshal(json_data, &restored_thing)
			Expect(err).To(BeNil())
		})

		It("wont format structs missing from the type store", func() {
			_, err := Format(thingy, &type_store)
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("FormatCapsule", func() {

		It("formats capsules correctly", func() {
			capsule_bytes, err := FormatCapsule(thingy, &populated_type_store, 1)
			Expect(err).To(BeNil())
			type_bytes := capsule_bytes[:2]
			size_bytes := capsule_bytes[2:6]
			capsule_data := capsule_bytes[6:]
			type_int := binary.LittleEndian.Uint16(type_bytes)
			size_int := binary.LittleEndian.Uint32(size_bytes)
			Expect(type_int).To(Equal(uint16(0)))
			Expect(size_int).To(Equal(uint32(len(capsule_data))))
			restored_capsule := &Capsule{}
			err = json.Unmarshal(capsule_data, &restored_capsule)
			Expect(err).To(BeNil())
			Expect(restored_capsule.RequestID).To(Equal(uint16(1)))
			Expect(restored_capsule.Type).To(Equal(uint16(1)))
			restored_thing := &Thingy{}
			err = json.Unmarshal([]byte(restored_capsule.Data), &restored_thing)
			Expect(err).To(BeNil())
			Expect(restored_thing.Name).To(Equal("test"))
			Expect(restored_thing.ID).To(Equal(1))
		})

		It("wont format a capsule with structs missing from the type store", func() {
			_, err := FormatCapsule(thingy, &type_store, 1)
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("NextStruct", func() {

		It("can read multiple structs", func() {
			sockets := make(chan net.Conn, 1)
			server, err := net.Listen("tcp", "localhost:5002")
			Expect(err).To(BeNil())
			defer server.Close()
			go func() {
				conn, _ := server.Accept()
				sockets <- conn
			}()
			client, err := net.Dial("tcp", "localhost:5002")
			Expect(err).To(BeNil())
			defer client.Close()
			server_side := <- sockets
			unicode_thingy := Thingy {
				Name:   "😃",
				ID:	2,
			}
			thingy_bytes, _ := Format(thingy, &populated_type_store)
			unicode_thingy_bytes, _ := Format(unicode_thingy, &populated_type_store)
			server_side.Write(thingy_bytes)
			server_side.Write(unicode_thingy_bytes)
			iface, err := NextStruct(client, &populated_type_store)
			Expect(err).To(BeNil())
			if restored_thingy, correct_type :=  iface.(*Thingy); correct_type {
				Expect(restored_thingy.Name).To(Equal(thingy.Name))
				Expect(restored_thingy.ID).To(Equal(thingy.ID))
			} else {
				Expect(correct_type).To(Equal(true))
			}
			iface, err = NextStruct(client, &populated_type_store)
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
			server, err := net.Listen("tcp", "localhost:5003")
			Expect(err).To(BeNil())
			defer server.Close()
			go func() {
				conn, _ := server.Accept()
				sockets <- conn
			}()
			client, err := net.Dial("tcp", "localhost:5003")
			Expect(err).To(BeNil())
			client.Close()
			_, err = NextStruct(client, &type_store)
			Expect(err).ToNot(BeNil())
		})

		It("returns nill when the struct is missing from the type store", func() {
			sockets := make(chan net.Conn, 1)
			server, err := net.Listen("tcp", "localhost:5004")
			Expect(err).To(BeNil())
			defer server.Close()
			go func() {
				conn, _ := server.Accept()
				sockets <- conn
			}()
			client, err := net.Dial("tcp", "localhost:5004")
			Expect(err).To(BeNil())
			defer client.Close()
			server_side := <- sockets
			thingy_bytes, _ := Format(thingy, &populated_type_store)
			server_side.Write(thingy_bytes)
			iface, err := NextStruct(client, &type_store)
			Expect(iface).To(BeNil())
			Expect(err).To(BeNil())
		})
	})
})


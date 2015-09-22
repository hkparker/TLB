package tlj_test

import (
	. "github.com/hkparker/TLJ"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
        "encoding/json"
	"reflect"
)
/*
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
*/

var _ = Describe("Common", func() {

	var (
		type_store TypeStore
		capsule Capsule
		thingy Thingy
	)

	BeforeEach(func() {
       	        type_store = NewTypeStore()
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

		It("should contain a functional captule builder when created", func() {
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
			inst_type := reflect.TypeOf(Thingy{})
			ptr_type := reflect.TypeOf(&Thingy{})
 			type_store.AddType(inst_type, ptr_type, BuildThingy)	
			thingy_bytes, _ := json.Marshal(thingy)
			iface := type_store.BuildType(1, thingy_bytes)
			if restored, correct_type := iface.(*Thingy); correct_type {
				Expect(restored.Name).To(Equal(thingy.Name))
				Expect(restored.ID).To(Equal(thingy.ID))
			} else {
				Expect(correct_type).To(Equal(true))
			}
		})
/*
		It("", func() {
			//
		})
*/
	})
})

package tlj

import (
//	"net"
	"reflect"
//	"errors"
	"encoding/json"
//	"encoding/binary"
	"testing"
)

type Thingy struct {
	Name	string
	ID		int
}

func BuildThingy(data []byte) interface{} {
		thing := &Thingy{}
		err := json.Unmarshal(data, &thing)
		if err != nil { return nil }
		return thing
}

func TestTypeStoreIsCorrectType(t *testing.T) {
	type_store := NewTypeStore()
	if reflect.TypeOf(type_store) != reflect.TypeOf(TypeStore{}) {
		t.Errorf("return value of NewTypeStore() != tlj.TypeStore")
	} 
}

func TestTypeStoreHasCapsuleBuilder(t *testing.T) {
	type_store := NewTypeStore()
	cap := Capsule {
		RequestID:	1,
		Type:		1,
		Data:		"test",
	}
	cap_bytes, _ := json.Marshal(cap)
	iface := type_store.BuildType(0, cap_bytes)
	if restored, ok := iface.(*Capsule); ok {
		if restored.RequestID != cap.RequestID {
			t.Errorf("capsule builder did not restore RequestID")
		}
		if restored.Type != cap.Type {
			t.Errorf("capsule builder did not restore Type")
		}
		if restored.Data != cap.Data {
			t.Errorf("capsule builder did not restore Data")
		}
	} else {
		t.Errorf("could not assert *Cpasule type on restored interface")
	}
}

func TestTypeStoreCanAddType(t *testing.T) {
	type_store := NewTypeStore()
	thingy_type := reflect.TypeOf(Thingy{})
	type_store.AddType(thingy_type, BuildThingy)
	if type_store.TypeCodes[thingy_type] != 1 {
		t.Errorf("call to AddType on new TypeStore did not create type_id of 1")
	}
}

func TestTypeStoreCanLookupCode(t *testing.T) {
	type_store := NewTypeStore()
	code, present := type_store.LookupCode(reflect.TypeOf(Capsule{}))
	if code != 0 || !present {
		t.Errorf("unable to lookup type_code for Capsule")
	}
}

func TestTypeStoreWontLookupBadCode(t *testing.T) {
	type_store := NewTypeStore()
	_, present := type_store.LookupCode(reflect.TypeOf(Thingy{}))
	if present {
		t.Errorf("nonexistent type returns a code")
	}
}

func TestTypeStoreCanBuildType(t *testing.T) {
	type_store := NewTypeStore()
	thingy_type := reflect.TypeOf(Thingy{})
	type_store.AddType(thingy_type, BuildThingy)
	if type_store.TypeCodes[thingy_type] != 1 {
		t.Errorf("call to AddType on new TypeStore did not create type_id of 1")
	}
	thingy := Thingy {
		Name:	"test",
		ID:		1,
	}
	marshalled, err := json.Marshal(thingy)
	if err != nil {
		t.Errorf("marshalling thingy returned an error")
	}
	iface := type_store.BuildType(1, marshalled)
	if restored, ok := iface.(*Thingy); ok {
		if restored.Name != thingy.Name {
			t.Errorf("string not presevered when building from marshalled struct")
		}
		if restored.ID != thingy.ID {
			t.Errorf("int not presevered when building from marshalled struct")
		}
	} else {
		t.Errorf("could not assert *Thingy type on restored interface")
	}
}

func TestTypeStoreWontBuildBadType(t *testing.T) {
	type_store := NewTypeStore()
	iface := type_store.BuildType(1, make([]byte, 0))
	if iface != nil {
		t.Errorf("type_Store built something with a nonexistent id")
	}
}

func TestTypeStoreWontBuildUnformattedData(t *testing.T) {
	type_store := NewTypeStore()
	iface := type_store.BuildType(0, []byte("notjson"))
	if iface != nil {
		t.Errorf("type_Store built something when bad data was supplied")
	}
}

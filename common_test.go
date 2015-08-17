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
/*
func TestTypeStoreHasCapsuleBuilder(t *testing.T) {
	type_store := NewTypeStore()
	// build a capsule
}
*/
func TestTypeStoreCanAddType(t *testing.T) {
	type_store := NewTypeStore()
	thingy_type := reflect.TypeOf(Thingy{})
	type_store.AddType(thingy_type, BuildThingy)
	if type_store.TypeCodes[thingy_type] != 1 {
		t.Errorf("call to AddType on new TypeStore did not create type_id of 1")
	}
}

//func TestTypeStoreWontAddBadFunc(t *testing.T) {
	//type_store := NewTypeStore()
	
//}

//func TestTypeStoreCanLookupCode(t *testing.T) {
	//type_store := NewTypeStore()
	
//}

//func TestTypeStoreWontLookupBadCode(t *testing.T) {
	//type_store := NewTypeStore()
	
//}
/*
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
	restored := type_store.BuildType(1, marshalled)
	if restored.Name != thingy.Name {
		t.Errorf("string not presevered when building from marshalled struct")
	}
	if restored.ID != thingy.ID {
		t.Errorf("int not presevered when building from marshalled struct")
	}
}
*/
//func TestTypeStoreWontBuildBadType(t *testing.T) {
	//type_store := NewTypeStore()
	
//}

//func TestTypeStoreWontBuildUnformattedData(t *testing.T) {
	//type_store := NewTypeStore()
	
//}

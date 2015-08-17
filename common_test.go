package tlj

import (
//	"net"
	"reflect"
//	"errors"
//	"encoding/json"
//	"encoding/binary"
	"testing"
)

func TestNewTypeStoreIsCorrectType(t *testing.T) {
	type_store := NewTypeStore()
	if reflect.TypeOf(type_store) != reflect.TypeOf(TypeStore{}) {
		t.Errorf("return value of NewTypeStore() != tlj.TypeStore")
	} 
}

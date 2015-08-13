package tlj

import (
	"reflect"
	"errors"
	"encoding/json"
	"encoding/binary"
)

type Capsule struct {
	RequestID	uint16
	Type		uint16
	Data		string
}

type TypeStore struct {
	Types			map[uint16]func()
	TypeCodes		map[reflect.Type]uint16
}

func NewTypeStore() TypeStore {
	type_store := TypeStore {
		Types:		make(map[uint16]func()),
		TypeCodes:	make(map[reflect.Type]uint16)
	}
	// add Request, Capsule, Responder, etc to the TypeStore
	return type_store
}

func (store *TypeStore) AddType(builder func([]byte)) error {
	// generate a new uint16 to represent this func
	// resolve the uint16 to the func
	// resolve the reflect.Type to the uint16
}

func (store *TypeStore) LookupCode(struct_type reflect.Type) uint16 {
	code, present := store.TypeCodes[struct_type]
	if !present { return nil } // needed?
	return code
}

func (store *TypeStore) BuildType(struct_code uint16, data []byte) interface{} {
	function, present := store.Types[struct_code]
	if !present { return nil } // needed?
	return function(data)
	
}

func (speaker interface{}) format(instance interface{}) ([]byte, error) {
	bytes, err := json.Marshal(instance)
	if err != nil { return nil, err }
	
	type_bytes := make([]byte, 2)
	struct_type := speaker.LookupCode(reflect.TypeOf(instance))
	if struct_type == nil { return nil, errors.New("cannot format unknown type") }
	binary.LittleEndian.PutUint16(type_bytes, struct_type)
	
	length := len(bytes)
	length_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(length_bytes, uint32(length))
	
	bytes = append(type_bytes, append(length_bytes, bytes...)...)

	return bytes, err
}

func (server Server) formatCapsule(instance interface{}, request_id uint16) ([]byte, error) {
	bytes, err := json.Marshal(instance)
	if err != nil { return bytes, err }

	struct_type, present := server.TypeStore.LookupCode([reflect.TypeOf(instance)])
	if !present {
		return nil, errors.New("struct type missing from TypeCodes")
	}
	
	resp := Capsule {
		RequestID:	request_id,
		Type:		struct_type,
		Data:		bytes
	}

	return format(resp)
}

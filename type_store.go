package tlj

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"sync"
)

type Capsule struct {
	RequestID uint16
	Type      uint16
	Data      string
}

type Builder func([]byte, TLJContext) interface{}

type TypeStore struct {
	Types      map[uint16]Builder
	TypeCodes  map[reflect.Type]uint16
	NextID     uint16
	InsertType *sync.Mutex
}

func NewTypeStore() TypeStore {
	type_store := TypeStore{
		Types:      make(map[uint16]Builder),
		TypeCodes:  make(map[reflect.Type]uint16),
		NextID:     1,
		InsertType: &sync.Mutex{},
	}

	capsule_builder := func(data []byte, _ TLJContext) interface{} {
		capsule := &Capsule{}
		err := json.Unmarshal(data, &capsule)
		if err != nil {
			return nil
		}
		return capsule
	}
	type_store.Types[0] = capsule_builder
	type_store.TypeCodes[reflect.TypeOf(Capsule{})] = 0
	type_store.TypeCodes[reflect.TypeOf(&Capsule{})] = 0

	return type_store
}

func (store *TypeStore) AddType(inst_type reflect.Type, ptr_type reflect.Type, builder Builder) error {
	if inst_type == nil {
		return errors.New("instance type cannot be nil")
	}
	if ptr_type == nil {
		return errors.New("pointer type cannot be nil")
	}
	if builder == nil {
		return errors.New("builder cannot be nil")
	}
	type_id := store.NextID
	store.NextID = store.NextID + 1
	store.InsertType.Lock()
	store.Types[type_id] = builder
	store.TypeCodes[inst_type] = type_id
	store.TypeCodes[ptr_type] = type_id
	store.InsertType.Unlock()
	return nil
}

func (store *TypeStore) LookupCode(struct_type reflect.Type) (uint16, bool) {
	val, present := store.TypeCodes[struct_type]
	return val, present
}

func (store *TypeStore) BuildType(struct_code uint16, data []byte, context TLJContext) interface{} {
	function, present := store.Types[struct_code]
	if !present {
		return nil
	}
	return function(data, context)
}

func (store *TypeStore) Format(instance interface{}) ([]byte, error) {
	bytes, err := json.Marshal(instance)
	if err != nil {
		return nil, err
	}

	type_bytes := make([]byte, 2)
	struct_type, present := store.LookupCode(reflect.TypeOf(instance))
	if !present {
		return type_bytes, errors.New("struct type missing from TypeStore")
	}
	binary.LittleEndian.PutUint16(type_bytes, struct_type)

	length := len(bytes)
	length_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(length_bytes, uint32(length))

	bytes = append(type_bytes, append(length_bytes, bytes...)...)

	return bytes, err
}

func (store *TypeStore) FormatCapsule(instance interface{}, request_id uint16) ([]byte, error) {
	bytes, err := json.Marshal(instance)
	if err != nil {
		return bytes, err
	}

	struct_type, present := store.LookupCode(reflect.TypeOf(instance))
	if !present {
		return bytes, errors.New("struct type missing from TypeStore")
	}

	capsule := Capsule{
		RequestID: request_id,
		Type:      struct_type,
		Data:      string(bytes),
	}

	return store.Format(capsule)
}

func (store *TypeStore) NextStruct(socket net.Conn, context TLJContext) (interface{}, error) {
	header := make([]byte, 6)
	n, err := socket.Read(header)
	if err != nil {
		return nil, err
	}
	if n != 6 {
		return nil, errors.New("invalid header")
	}

	type_bytes := header[:2]
	size_bytes := header[2:]

	type_int := binary.LittleEndian.Uint16(type_bytes)
	size_int := binary.LittleEndian.Uint32(size_bytes)

	struct_data := make([]byte, size_int)
	_, err = socket.Read(struct_data)
	if err != nil {
		return nil, err
	}

	recieved_struct := store.BuildType(type_int, struct_data, context)

	return recieved_struct, nil
}

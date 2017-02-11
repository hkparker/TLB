package tlj

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"reflect"
	"sync"
)

//
// A capsule is used to maintain a session between a client
// and a server when using sever.AcceptRequest, client.Request,
// or request.OnResponse.
//
type Capsule struct {
	RequestID uint16
	Type      uint16
	Data      string
}

//
// Builders are functions that take the raw payload in the TLV
// protocol and parse the JSON and run any other validations
// that may be nessicary based on the context before returning
// the struct.
//
type Builder func([]byte, TLJContext) interface{}

//
// A TypeStore stores the information needed to marshal,
// unmsrahsl, and recognize all types passed between all
// other instances of TLJ that are communicated with.
//
type TypeStore struct {
	Types      map[uint16]Builder
	TypeCodes  map[reflect.Type]uint16
	NextID     uint16
	InsertType *sync.Mutex
}

//
// Create a new TypeStore with type 0 being a Capsule.
//
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

//
// Insert a new type into the TypeStore by providing a reflect.Type
// of the struct and a pointer to the struct, as well as the Builder
// that will be used to construct the type.  Types must be JSON
// serializable.
//
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

//
// Given the reflect.Type of a type in the TypeStore, return
// the uint16 that is used to identify this type over the
// network, and a boolean to indicate if the type was present
// in the TypeStore.
//
func (store *TypeStore) LookupCode(struct_type reflect.Type) (uint16, bool) {
	val, present := store.TypeCodes[struct_type]
	return val, present
}

//
// Call the Builder function for a given type on some data if
// the type exists in the type store, return nil if the type
// does not exist.
//
func (store *TypeStore) BuildType(struct_code uint16, data []byte, context TLJContext) interface{} {
	function, present := store.Types[struct_code]
	if !present {
		return nil
	}
	return function(data, context)
}

//
// Take any JSON serializable struct that is in the TypeStore
// and return the byte sequence to send on the network to deliver
// the struct to the other instance of TLJ, as well as any errors.
//
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

//
// Take a struct and format it inside of a Capsule so it can
// be sent statefully to another TLJ instance.
//
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

//
// Read a struct from a net.Conn interface using the types contained
// in a TypeStore.
//
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

	if _, present := store.Types[type_int]; !present {
		return nil, errors.New("type code on received struct not in type store")
	}

	struct_data := make([]byte, 0)
	total_read := 0
	for total_read < int(size_int) {
		buf := make([]byte, int(size_int)-total_read)
		n, err = socket.Read(buf)
		total_read += n
		if err == io.EOF {
			if total_read != int(size_int) {
				return nil, errors.New("EOF before all data read")
			}
		} else if err != nil {
			return nil, err
		}
		struct_data = append(struct_data, buf[:n]...)
	}

	recieved_struct := store.BuildType(type_int, struct_data, context)

	return recieved_struct, nil
}

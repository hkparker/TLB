package tlj

import (
	"reflect"
	"encoding/json"
	"encoding/binary"
)

type Speaker interface {
	// has .Types, .TypeCodes
}

func (client Speaker) format(instance interface{}) ([]byte, error) {
	// Get value
	bytes, err := json.Marshal(instance)
	if err != nil { return bytes, err }
	
	// Get type
	type_bytes := make([]byte, 2)
	struct_type := client.TypeCodes[reflect.TypeOf(instance)]
	binary.LittleEndian.PutUint16(type_bytes, struct_type)
	
	// Get length
	length := len(bytes)
	length_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(length_bytes, uint32(length))
	
	// Merge everything
	bytes = append(type_bytes, append(length_bytes, bytes...)...)

	return bytes, err
}

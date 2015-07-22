package tlj

import (
	"reflect"
	"encoding/json"
	"encoding/binary"
)

type Response struct {
	RequestID	uint16
	Type		uint16
	Data		string
}

type Speaker interface {
	// has .Types, .TypeCodes
}

func (speaker Speaker) format(instance interface{}) ([]byte, error) {
	// Get value
	bytes, err := json.Marshal(instance)
	if err != nil { return bytes, err }
	
	// Get type
	type_bytes := make([]byte, 2)
	struct_type := speaker.TypeCodes[reflect.TypeOf(instance)]		// if nil?
	binary.LittleEndian.PutUint16(type_bytes, struct_type)
	
	// Get length
	length := len(bytes)
	length_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(length_bytes, uint32(length))
	
	// Merge everything
	bytes = append(type_bytes, append(length_bytes, bytes...)...)

	return bytes, err
}


func (server Server) formatResponse(instance interface{}, request_id uint16) ([]byte, error) {
	// marshal the instance
	// lookup its type in server's map
	// create a response struct from those + request_id
	// format
}

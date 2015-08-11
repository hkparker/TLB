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

func (speaker interface{}) format(instance interface{}) ([]byte, error) {
	// Get value
	bytes, err := json.Marshal(instance)
	if err != nil { return bytes, err }
	
	// Get type
	type_bytes := make([]byte, 2)
	struct_type := speaker.TypeCodes[reflect.TypeOf(instance)]		// if nil?
	binary.LittleEndian.PutUint16(type_bytes, struct_type)			// dont actually need this because map stores type uint16?
	
	// Get length
	length := len(bytes)
	length_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(length_bytes, uint32(length))
	
	// Merge everything
	bytes = append(type_bytes, append(length_bytes, bytes...)...)

	return bytes, err
}

func (server Server) formatCapsule(instance interface{}, request_id uint16) ([]byte, error) {
	bytes, err := json.Marshal(instance)
	if err != nil { return bytes, err }

	struct_type, present := server.TypeCodes[reflect.TypeOf(instance)]
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

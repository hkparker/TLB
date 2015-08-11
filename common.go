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
	binary.LittleEndian.PutUint16(type_bytes, struct_type)		// dont actually need this because map stores type uint16?
	
	// Get length
	length := len(bytes)
	length_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(length_bytes, uint32(length))
	
	// Merge everything
	bytes = append(type_bytes, append(length_bytes, bytes...)...)

	return bytes, err
}


func (server Server) formatResponse(instance interface{}, request_id uint16) ([]byte, error) {		// is this still needed now that I am creating a Response struct?
	bytes, err := json.Marshal(instance)
	if err != nil { return bytes, err }

	struct_type := server.TypeCodes[reflect.TypeOf(instance)]		// if nil?
	
	resp := Response {
		RequestID:	request_id,
		Type:		struct_type,
		Data:		bytes
	}

	return format(resp)
}

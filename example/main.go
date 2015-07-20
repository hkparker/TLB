package main

import (
	"fmt"
	"../"
	"reflect"
	"encoding/json"
	"encoding/binary"
)

type Hayden struct {
	Name string
}

func format(instance interface{}, types map[reflect.Type]uint16) ([]byte, error) {

	// Get type
	type_bytes := make([]byte, 2)
	struct_type := types[reflect.TypeOf(instance)]
	binary.LittleEndian.PutUint16(type_bytes, struct_type)

	// Get value
	bytes, err := json.Marshal(instance)
	if err != nil { return bytes, err }

	// Get length
	length := len(bytes)
	length_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(length_bytes, uint32(length))

	// Group everything together
	bytes = append(type_bytes, append(length_bytes, bytes...)...)

	return bytes, err
}

func main() {
	//types := make(map[reflect.Type]uint16)
	//types[reflect.TypeOf(Hayden{})] = 1
	types := map[reflect.Type]uint16 {
		reflect.TypeOf(Hayden{}): 1,
	}
	h := Hayden{Name: "Hayden"}
	fmt.Println(format(h, types))
	//fmt.Println(reflect.TypeOf(Hayden{}))
}


package tlb_test

import (
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func BenchmarkBSON(b *testing.B) {
	thingy := Thingy{
		Name: string(make([]byte, 16000)),
		ID:   1,
	}
	for n := 0; n < b.N; n++ {
		data, _ := bson.Marshal(thingy)
		bson.Unmarshal(data, &Thingy{})
	}
}

func BenchmarkJSON(b *testing.B) {
	thingy := Thingy{
		Name: string(make([]byte, 16000)),
		ID:   1,
	}
	for n := 0; n < b.N; n++ {
		data, _ := json.Marshal(thingy)
		json.Unmarshal(data, &Thingy{})
	}
}

package marshal

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type JSONPb struct {
	protojson.MarshalOptions
	protojson.UnmarshalOptions
}

func (j *JSONPb) Marshal(v interface{}) ([]byte, error) {
	if pb, ok := v.(proto.Message); ok {
		return j.MarshalOptions.Marshal(pb)
	}
	return json.Marshal(v)
}

func (j *JSONPb) Unmarshal(d []byte, v interface{}) error {
	if pb, ok := v.(proto.Message); ok {
		return j.UnmarshalOptions.Unmarshal(d, pb)
	}
	return json.Unmarshal(d, v)
}

func (j *JSONPb) String() string {
	return "jsonpb"
}

func (*JSONPb) ContentType(_ interface{}) string {
	return "application/json"
}

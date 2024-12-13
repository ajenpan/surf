package marshal

import "google.golang.org/protobuf/encoding/protojson"

type Marshaler interface {
	Id() uint8
	ContentType() string
	// Marshal marshals "v" into byte sequence.
	Marshal(v interface{}) ([]byte, error)
	// Unmarshal unmarshals "data" into "v".
	// "v" must be a pointer value.
	Unmarshal(data []byte, v interface{}) error
}

type MarshalerId = uint8

const MarshalerId_protobuf MarshalerId = 0
const MarshalerId_json MarshalerId = 1
const MarshalerId_invalid MarshalerId = 255

func NewMarshaler(typ MarshalerId) Marshaler {
	switch typ {
	case MarshalerId_protobuf:
		return &ProtoMarshaler{}
	case MarshalerId_json:
		return &JSONPb{
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
		}
	}
	return nil
}

func NameToId(typ string) MarshalerId {
	switch typ {
	case "application/json":
		return MarshalerId_json
	case "application/protobuf":
		return MarshalerId_protobuf
	}
	return MarshalerId_invalid
}

func IdToName(id MarshalerId) string {
	marshal := NewMarshaler(id)
	if marshal == nil {
		return "invalid_marshale_name"
	}
	return marshal.ContentType()
}

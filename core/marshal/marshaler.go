package marshal

type Marshaler interface {
	// Marshal marshals "v" into byte sequence.
	Marshal(v interface{}) ([]byte, error)
	// Unmarshal unmarshals "data" into "v".
	// "v" must be a pointer value.
	Unmarshal(data []byte, v interface{}) error
	// ContentType returns the Content-Type which this marshaler is responsible for.
	// The parameter describes the type which is being marshalled, which can sometimes
	// affect the content type returned.
	ContentType(interface{}) string
	ContentTypeID() uint8
}

func NewMarshalerById(typ uint8) Marshaler {
	switch typ {
	case 0:
		return &ProtoMarshaler{}
	case 1:
		return &JSONPb{}
	}
	return nil
}

func NewMarshalerByName(typ string) Marshaler {
	switch typ {
	case "application/json":
		return &JSONPb{}
	case "application/protobuf":
		return &ProtoMarshaler{}
	}
	return nil
}

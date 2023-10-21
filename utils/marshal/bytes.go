package marshal

import (
	"errors"
)

type BytesMarshaler struct{}

var ErrInvalidMessage = errors.New("invalid message")

func (n BytesMarshaler) Marshal(v interface{}) ([]byte, error) {
	switch ve := v.(type) {
	case *[]byte:
		return *ve, nil
	case []byte:
		return ve, nil
	}
	return nil, ErrInvalidMessage
}

func (n BytesMarshaler) Unmarshal(d []byte, v interface{}) error {
	switch ve := v.(type) {
	case *[]byte:
		*ve = d
	}
	return ErrInvalidMessage
}

func (n BytesMarshaler) String() string {
	return "bytes"
}

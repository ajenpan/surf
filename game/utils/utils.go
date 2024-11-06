package utils

import (
	"errors"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func GetMessageMsgID(msg protoreflect.MessageDescriptor) (uint32, error) {
	MSGIDDesc := msg.Enums().ByName("MSGID")
	if MSGIDDesc == nil {
		return 0, errors.New("MSGID Desc is nil")
	}
	IDDesc := MSGIDDesc.Values().ByName("ID")
	if IDDesc == nil {
		return 0, errors.New("ID Desc is nil")
	}
	return uint32(IDDesc.Number()), nil
}

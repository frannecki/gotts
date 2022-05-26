package codec

import (
	"testing"
)

func compareProtoTestMessage(msg1 *TestMessage, msg2 *TestMessage) bool {
	return msg1.Key == msg2.Key && msg1.Value == msg2.Value && msg1.Id == msg2.Id
}

func TestMarshalUnmarshal(t *testing.T) {
	msg := TestMessage{Key: "re", Value: "ply", Id: 4}
	codec := NewProtobufCodec()
	codec.registerProtoMessage(msg.ProtoReflect().Type())
	marshaled := marshalProtoMessage(&msg)
	msg1, err := codec.unmarshalProtoMessage(marshaled)
	if err != nil {
		t.Error(err)
	}
	if !compareProtoTestMessage(&msg, msg1.(*TestMessage)) {
		t.Errorf("proto message inconsistent. Brfore marshaling: %v. After marshaling: %v", &msg, &msg1)
	}
}

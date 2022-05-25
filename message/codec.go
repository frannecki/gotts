package message

import (
	"bytes"
	"encoding/binary"
	"log"

	"github.com/frannecki/gotts/common"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type CodecCallback func(*common.Channel, interface{})

type Codec interface {
	Send(*common.Channel, interface{})
	Recv(*common.Channel, *common.Buffer)
	RegisterCallback(interface{}, CodecCallback)
	RegisterDefaultCallback(CodecCallback)
}

type unsupportedProtoMessageError struct{}

func (e *unsupportedProtoMessageError) Error() string {
	return "no callback registered for such proto message type"
}

type ProtobufCodec struct {
	callbacks       map[protoreflect.MessageType]CodecCallback
	prototypes      map[string]protoreflect.MessageType
	defaultCallback CodecCallback
}

func marshalProtoMessage(msg proto.Message) []byte {
	marshaled, err := proto.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}
	name := msg.ProtoReflect().Descriptor().Name()
	name_len := uint16(len(name))
	marshaled_len := uint16(len(marshaled))
	// marshaled message length + name length + name header length + header length
	msg_len := marshaled_len + name_len + 4
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, msg_len)
	binary.Write(buf, binary.BigEndian, name_len)
	binary.Write(buf, binary.BigEndian, []byte(name))
	binary.Write(buf, binary.BigEndian, marshaled)
	return buf.Bytes()
}

func (codec *ProtobufCodec) unmarshalProtoMessage(msg_buf []byte) (proto.Message, error) {
	name_header := msg_buf[2:4] // uint16 header
	name_len := binary.BigEndian.Uint16(name_header)
	prototype, ok := codec.prototypes[string(msg_buf[4:(4+name_len)])]
	if !ok {
		return nil, &unsupportedProtoMessageError{}
	}
	marshaled := msg_buf[(4 + name_len):]
	msg := prototype.New().Interface()
	if err := proto.Unmarshal(marshaled, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (codec *ProtobufCodec) Send(channel *common.Channel, msg proto.Message) {
	channel.Send(marshalProtoMessage(msg))
}

func (codec *ProtobufCodec) Recv(channel *common.Channel, buffer *common.Buffer) {
	for {
		header := buffer.Peek(2) // uint16 header
		if len(header) < 2 {
			break
		}
		msg_len := binary.BigEndian.Uint16(header)
		if msg_len < 4 {
			log.Fatal("proto message length < 4")
		}
		if buffer.ReadableBytes() < msg_len {
			// a message should contain at least a header and a name header
			break
		}
		msg_buf := buffer.Peek(msg_len)
		if msg, err := codec.unmarshalProtoMessage(msg_buf); err == nil {
			buffer.HaveRead(msg_len)
			codec.callbacks[msg.ProtoReflect().Type()](channel, msg)
		} else if _, ok := err.(*unsupportedProtoMessageError); ok {
			buffer.HaveRead(msg_len)
			codec.defaultCallback(channel, msg_buf)
		} else {
			break
		}
	}
}

func (codec *ProtobufCodec) registerProtoMessage(prototype protoreflect.MessageType) {
	codec.prototypes[string(prototype.Descriptor().Name())] = prototype
}

func (codec *ProtobufCodec) RegisterCallback(prototype protoreflect.MessageType, callback CodecCallback) {
	codec.callbacks[prototype] = callback
	codec.registerProtoMessage(prototype)
}

func (codec *ProtobufCodec) RegisterDefaultCallback(callback CodecCallback) {
	codec.defaultCallback = callback
}

func NewProtobufCodec() *ProtobufCodec {
	return &ProtobufCodec{
		callbacks:  make(map[protoreflect.MessageType]CodecCallback),
		prototypes: make(map[string]protoreflect.MessageType),
	}
}

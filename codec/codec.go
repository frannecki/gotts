package codec

import (
	"bytes"
	"encoding/binary"
	"log"

	"crypto/md5"

	"github.com/frannecki/gotts/common"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	proto_header_length       = 2
	proto_name_header_length  = 2
	proto_checksum_length     = 16
	proto_total_header_length = proto_header_length + proto_name_header_length
	proto_min_message_length  = proto_total_header_length + proto_checksum_length
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

type dataIntegrityError struct{}

func (e *dataIntegrityError) Error() string {
	return "Data not consistent with checksum"
}

type ProtobufCodec struct {
	callbacks       map[protoreflect.MessageType]CodecCallback
	prototypes      map[string]protoreflect.MessageType
	defaultCallback CodecCallback
}

func encode(msg proto.Message) []byte {
	marshaled, err := proto.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}
	name := msg.ProtoReflect().Descriptor().Name()
	name_len := uint16(len(name))
	marshaled_len := uint16(len(marshaled))
	// marshaled message length + name length + name header length + header length + checksum length
	msg_len := marshaled_len + name_len + proto_min_message_length
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, msg_len)
	binary.Write(buf, binary.BigEndian, name_len)
	binary.Write(buf, binary.BigEndian, []byte(name))
	binary.Write(buf, binary.BigEndian, marshaled)
	checksum := md5.Sum(buf.Bytes())
	binary.Write(buf, binary.BigEndian, checksum)
	return buf.Bytes()
}

func (codec *ProtobufCodec) decode(msg_buf []byte) (proto.Message, error) {
	msg_len := len(msg_buf)
	checksum := msg_buf[(msg_len - proto_checksum_length):]
	checksum_expected := md5.Sum(msg_buf[:(msg_len - proto_checksum_length)])
	if !bytes.Equal(checksum, checksum_expected[:]) {
		return nil, &dataIntegrityError{}
	}

	name_header := msg_buf[proto_header_length:proto_total_header_length] // uint16 name header
	name_len := binary.BigEndian.Uint16(name_header)
	prototype, ok := codec.prototypes[string(msg_buf[proto_total_header_length:(proto_total_header_length+name_len)])]
	if !ok {
		return nil, &unsupportedProtoMessageError{}
	}
	marshaled := msg_buf[(proto_total_header_length + name_len):(msg_len - proto_checksum_length)]
	msg := prototype.New().Interface()
	if err := proto.Unmarshal(marshaled, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (codec *ProtobufCodec) Send(channel *common.Channel, msg proto.Message) {
	channel.Send(encode(msg))
}

func (codec *ProtobufCodec) Recv(channel *common.Channel, buffer *common.Buffer) {
	for {
		header := buffer.Peek(proto_header_length) // uint16 header
		if len(header) < proto_header_length {
			break
		}
		msg_len := binary.BigEndian.Uint16(header)
		if msg_len < proto_min_message_length {
			log.Fatal("proto message length <", proto_min_message_length)
		}
		if buffer.ReadableBytes() < msg_len {
			// a message should contain at least a header and a name header
			break
		}
		msg_buf := buffer.Peek(msg_len)
		if msg, err := codec.decode(msg_buf); err == nil {
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

package common

import (
	"log"
	"net"
)

type Buffer struct {
	buf        []byte
	readIndex  uint16
	writeIndex uint16
}

func (buffer *Buffer) Read(n uint16) []byte {
	peeked := buffer.Peek(n)
	nbytes := uint16(len(peeked))
	buffer.HaveRead(nbytes)
	return peeked
}

func (buffer *Buffer) Peek(n uint16) []byte {
	readable := buffer.ReadableBytes()
	if n > readable {
		n = readable
	}
	return buffer.buf[buffer.readIndex:(buffer.readIndex + n)]
}

func (buffer *Buffer) HaveRead(n uint16) {
	if n > buffer.ReadableBytes() {
		log.Fatal("not enough buffer to read")
	}
	buffer.readIndex += n
	if buffer.readIndex == buffer.writeIndex {
		buffer.readIndex = 0
		buffer.writeIndex = 0
	}
}

func (buffer *Buffer) ReadAll() []byte {
	return buffer.Read(buffer.ReadableBytes())
}

func (buffer *Buffer) Write(s []byte) {
	slen := uint16(len(s))
	buffer.buf = append(buffer.buf[:buffer.writeIndex], s...)
	buffer.writeIndex += slen
}

func (buffer *Buffer) ReadableBytes() uint16 {
	return buffer.writeIndex - buffer.readIndex
}

func (buffer *Buffer) ReadAsString(n uint16) string {
	return string(buffer.Read(n))
}

func (buffer *Buffer) ReadAllAsString() string {
	return string(buffer.ReadAll())
}

func (buffer *Buffer) WriteString(s string) {
	buffer.Write([]byte(s))
}

type Channel struct {
	Name       string
	Conn       net.Conn
	ReadBuffer Buffer
}

func (channel *Channel) Send(msg []byte) {
	nbytes := 0
	for nbytes < len(msg) {
		n, err := channel.Conn.Write(msg[nbytes:])
		if err != nil {
			// TODO: properly handle error
			log.Printf("Error for connection %s: %s\n", channel.Name, err.Error())
			return
		}
		nbytes += n
	}
}

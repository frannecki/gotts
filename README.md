# Tiny tcp server/client in go
## Usage
### Installation
```sh
go get github.com/frannecki/gotts
go get google.golang.org/protobuf
```

### TCP server
```go
// proto file messagetest/test.proto
// Run `protoc -I. --go_out=. messagetest/test.proto` to compile

///////////////////////////////////
// syntax = "proto3";

// package proto;
// option go_package = "messagetest/";

// message TestMessage
// {
//     string key = 1;
//     string value = 2;
//     bool success = 3;
//     uint32 id = 4;
//     bool validate = 5;
// }
///////////////////////////////////
package main

import (
	"fmt"
	"log"

	"gotts_server/messagetest"

	cdc "github.com/frannecki/gotts/codec"
	"github.com/frannecki/gotts/common"
	"github.com/frannecki/gotts/server"
)

func main() {
	fmt.Println("Example TCP Server in Golang")
	msg0 := messagetest.TestMessage{
		Key:      "hello",
		Value:    "good",
		Success:  true,
		Id:       45,
		Validate: false,
	}
	codec := cdc.NewProtobufCodec()
	codec.RegisterCallback(
		msg0.ProtoReflect().Type(),
		func(channel *common.Channel, msg interface{}) {
			if msg, ok := msg.(*messagetest.TestMessage); ok {
				fmt.Println(msg)
				msg.Id += 1
				codec.Send(channel, msg)
			} else {
				log.Println("Unknown message")
			}
		})
	// server := server.TcpServer{Callback: readCallback}
	server := server.NewTcpServer()
	server.RegisterReadCallback(
		func(channel *common.Channel, buffer *common.Buffer) {
			codec.Recv(channel, buffer)
		})
	server.RegisterConnectionCallback(func(channel *common.Channel) {
		codec.Send(channel, &msg0)
	})
	server.Serve("0.0.0.0", 8082)
}
```

### TCP client
```go
// proto file messagetest/test.proto
// Run `protoc -I. --go_out=. messagetest/test.proto` to compile

///////////////////////////////////
// syntax = "proto3";

// package proto;
// option go_package = "messagetest/";

// message TestMessage
// {
//     string key = 1;
//     string value = 2;
//     bool success = 3;
//     uint32 id = 4;
//     bool validate = 5;
// }
///////////////////////////////////
package main

import (
	"fmt"
	"log"

	"gotts_client/messagetest"

	"github.com/frannecki/gotts/client"
	cdc "github.com/frannecki/gotts/codec"
	"github.com/frannecki/gotts/common"
)

func main() {
	client := client.NewTcpClient()
	codec := cdc.NewProtobufCodec()
	msg0 := messagetest.TestMessage{
		Key:      "hello",
		Value:    "good",
		Success:  true,
		Id:       45,
		Validate: false,
	}
	codec.RegisterCallback(
		msg0.ProtoReflect().Type(),
		func(channel *common.Channel, msg interface{}) {
			if msg, ok := msg.(*messagetest.TestMessage); ok {
				fmt.Println(msg)
			} else {
				log.Println("Unknown message")
			}
		})
	client.RegisterReadCallback(
		func(channel *common.Channel, buffer *common.Buffer) {
			// buf := buffer.ReadAll()
			// channel.Send(buf)
			codec.Recv(channel, buffer)
		})
	if err := client.Connect("127.0.0.1", 8082); err != nil {
		log.Fatal(err)
	}
	client.Run()
}
```

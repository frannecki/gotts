# Tiny tcp server/client in go
## Usage
### Server
```go
package main

import (
	"fmt"
	"log"

	"github.com/frannecki/gotts/common"
	"github.com/frannecki/gotts/message"
	"github.com/frannecki/gotts/server"
)

func main() {
	msg0 := message.TestMessage{
		Key:      "hello",
		Value:    "good",
		Success:  true,
		Id:       45,
		Validate: false,
	}
	codec := message.NewProtobufCodec()
	codec.RegisterCallback(
		msg0.ProtoReflect().Type(),
		func(channel *common.Channel, msg interface{}) {
			if msg, ok := msg.(*message.TestMessage); ok {
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

### Client
```go
package main

import (
	"log"

	"github.com/frannecki/gotts/client"
	"github.com/frannecki/gotts/common"
)

func main() {
	client := client.NewTcpClient()
	client.RegisterReadCallback(
		func(channel *common.Channel, buffer *common.Buffer) {
			buf := buffer.ReadAll()
			channel.Send(buf)
		})
	if err := client.Connect("127.0.0.1", 8082); err != nil {
		log.Fatal(err)
	}
	client.Run()
}

```

package client

import (
	"fmt"
	"net"

	"github.com/frannecki/gotts/common"
)

type TcpClient struct {
	Channel            *common.Channel
	callback           func(*common.Channel, *common.Buffer)
	connectionCallback func(*common.Channel)
}

func (client *TcpClient) Connect(ip string, port uint16) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return err
	}
	name := conn.RemoteAddr().String()
	client.Channel = &common.Channel{Conn: conn, Name: name}
	return nil
}

func (client *TcpClient) RegisterReadCallback(callback func(*common.Channel, *common.Buffer)) {
	client.callback = callback
}

func (server *TcpClient) RegisterConnectionCallback(callback func(*common.Channel)) {
	server.connectionCallback = callback
}

func (client *TcpClient) Run() {
	if client.connectionCallback != nil {
		client.connectionCallback(client.Channel)
	}
	read_buf := make([]byte, 256)
	for {
		n, err := client.Channel.Conn.Read(read_buf)
		if err != nil {
			break
		}
		client.Channel.ReadBuffer.Write(read_buf[:n])
		if client.callback != nil {
			client.callback(client.Channel, &client.Channel.ReadBuffer)
		}
	}
}

func NewTcpClient() *TcpClient {
	return &TcpClient{}
}

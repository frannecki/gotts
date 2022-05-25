package server

import (
	"fmt"
	"log"
	"net"

	"github.com/frannecki/gotts/common"
)

type TcpServer struct {
	connections        map[string]*common.Channel
	callback           func(*common.Channel, *common.Buffer)
	channelEntering    chan *common.Channel
	channelExiting     chan *common.Channel
	groupMessage       chan string
	connectionCallback func(*common.Channel)
}

func (server *TcpServer) handleTcpConnection(conn net.Conn) {
	defer conn.Close()
	name := conn.RemoteAddr().String()
	channel := &common.Channel{Conn: conn, Name: name}
	server.channelEntering <- channel

	log.Println("Handling connection from", name)
	if server.connectionCallback != nil {
		server.connectionCallback(channel)
	}
	read_buf := make([]byte, 256)
	for {
		n, err := channel.Conn.Read(read_buf)
		if err != nil {
			break
		}
		channel.ReadBuffer.Write(read_buf[:n])
		if server.callback != nil {
			server.callback(channel, &channel.ReadBuffer)
		}
	}
	log.Printf("Connection from %s has disconnected\n", name)
	server.channelExiting <- channel
}

func (server *TcpServer) monitorTcpConnections() {
	for {
		select {
		case channel := <-server.channelEntering:
			server.connections[channel.Name] = channel
		case channel := <-server.channelExiting:
			delete(server.connections, channel.Name)
		}
	}
}

func (server *TcpServer) Serve(ip string, port uint16) {
	go server.monitorTcpConnections()
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error while accepting: ", err)
			continue
		}
		go server.handleTcpConnection(conn)
	}
}

func (server *TcpServer) RegisterReadCallback(callback func(*common.Channel, *common.Buffer)) {
	server.callback = callback
}

func (server *TcpServer) RegisterConnectionCallback(callback func(*common.Channel)) {
	server.connectionCallback = callback
}

func NewTcpServer() *TcpServer {
	return &TcpServer{
		connections:     make(map[string]*common.Channel),
		channelEntering: make(chan *common.Channel),
		channelExiting:  make(chan *common.Channel),
		groupMessage:    make(chan string),
	}
}

package server

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/frannecki/gotts/common"
)

type channelWrapper struct {
	channel *common.Channel
	closing chan struct{}
}

type TcpServer struct {
	connections        map[string]*channelWrapper
	callback           func(*common.Channel, *common.Buffer)
	channelEntering    chan *channelWrapper
	channelExiting     chan *channelWrapper
	groupMessage       chan string
	connectionCallback func(*common.Channel)
	done               chan struct{}
	connDone           chan struct{}
	connWg             sync.WaitGroup
}

func (server *TcpServer) handleTcpConnection(conn net.Conn) {
	defer server.connWg.Done()
	defer conn.Close()
	name := conn.RemoteAddr().String()
	channel := &common.Channel{Conn: conn, Name: name}
	chan_closing := make(chan struct{})
	channel_wrapper := &channelWrapper{channel: channel, closing: chan_closing}
	server.channelEntering <- channel_wrapper

	log.Println("Handling connection from", name)
	if server.connectionCallback != nil {
		server.connectionCallback(channel)
	}
	read_buf := make([]byte, 256)
	go func() {
		<-chan_closing
		channel.Conn.Close()
	}()
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
	server.channelExiting <- channel_wrapper
}

func (server *TcpServer) monitorTcpConnections() {
	for {
		select {
		case <-server.connDone:
			// close all connections
			for _, channel_wrapper := range server.connections {
				channel_wrapper.closing <- struct{}{}
			}
		case channel_wrapper := <-server.channelEntering:
			server.connections[channel_wrapper.channel.Name] = channel_wrapper
		case channel_wrapper := <-server.channelExiting:
			log.Println("Connection disconnected:", channel_wrapper.channel.Name)
			delete(server.connections, channel_wrapper.channel.Name)
		}
	}
}

func (server *TcpServer) Serve(ip string, port uint16) {
	go server.monitorTcpConnections()
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		<-server.done
		listener.Close()
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		server.connWg.Add(1)
		go server.handleTcpConnection(conn)
	}
	server.connDone <- struct{}{}
	server.connWg.Wait()
	log.Println("Server exiting...")
}

func (server *TcpServer) Stop() {
	server.done <- struct{}{}
}

func (server *TcpServer) RegisterReadCallback(callback func(*common.Channel, *common.Buffer)) {
	server.callback = callback
}

func (server *TcpServer) RegisterConnectionCallback(callback func(*common.Channel)) {
	server.connectionCallback = callback
}

func NewTcpServer() *TcpServer {
	return &TcpServer{
		connections:     make(map[string]*channelWrapper),
		channelEntering: make(chan *channelWrapper),
		channelExiting:  make(chan *channelWrapper),
		groupMessage:    make(chan string),
		done:            make(chan struct{}),
		connDone:        make(chan struct{}),
	}
}

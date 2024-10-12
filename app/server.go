package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
)

const ServerPort = 7007

type Server struct {
	port      uint16
	ipAddrStr string
	client    *Client
}

type Client struct {
	maxClients        uint8
	connectedClients  uint8
	clientConnections []net.Conn
	lock              sync.Mutex
}

func (s *Server) GetIpAddrStr() string {
	return s.ipAddrStr + ":" + strconv.FormatUint(uint64(s.port), 10)
}

func (s *Server) InitServer(maxClients uint8) {
	serverAddr, err := net.ResolveTCPAddr("tcp", s.GetIpAddrStr())
	if err != nil {
		fmt.Println("Failed to bind port to localhost")
		os.Exit(1)
	}

	listener, err := net.ListenTCP("tcp", serverAddr)
	if err != nil {
		fmt.Println("Failed to start server on localhost")
		os.Exit(1)
	}

	s.client = &Client{
		maxClients:        maxClients,
		connectedClients:  0,
		clientConnections: make([]net.Conn, maxClients),
	}

	fmt.Println("Listening to Connections")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}

		go s.handleClientConnection(conn)
	}
}

func (s *Server) handleClientConnection(conn net.Conn) {
	defer conn.Close()

	if !s.RegisterClient(conn) {
		conn.Write([]byte("Max number of registrations have been reached!\n"))
		return
	}

	fmt.Println("Registered Client")

	buffer := make([]byte, 512)

	for {
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(buffer[:bytesRead])

		if buffer[0] == 255 {
			fmt.Printf("Closing Conn to %s\n", conn.RemoteAddr().String())
			break
		}
		s.BroadcastMessage(conn, buffer[:bytesRead])
	}
}

func (s *Server) RegisterClient(conn net.Conn) bool {
	s.client.lock.Lock()

	s.client.connectedClients += 1
	didRegister := false
	for i := uint8(0); i < s.client.maxClients; i++ {
		if s.client.clientConnections[i] == nil {
			s.client.clientConnections[i] = conn
			didRegister = true
			break
		}
	}

	s.client.lock.Unlock()
	return didRegister
}

func (s *Server) BroadcastMessage(conn net.Conn, message []byte) {
	fmt.Printf("Broadcasting message from %s\n", conn.RemoteAddr().String())
	for i := uint8(0); i < s.client.maxClients; i++ {
		if s.client.clientConnections[i] == nil {
			continue
		}
		// if s.client.clientConnections[i].RemoteAddr() == conn.RemoteAddr() {
		// 	continue
		// }

		s.client.clientConnections[i].Write(message)
	}
}

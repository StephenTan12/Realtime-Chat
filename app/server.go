package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
)

type Server struct {
	port      uint16
	ipAddrStr string
	clients   *ClientsData
}

type ClientsData struct {
	maxClients        uint8
	connectedClients  uint8
	clientConnections []net.Conn
	lock              sync.RWMutex
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
	defer listener.Close()

	s.clients = &ClientsData{
		maxClients:        maxClients,
		connectedClients:  0,
		clientConnections: make([]net.Conn, maxClients),
	}

	fmt.Println("Listening to Connections")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}

		go s.handleClientConnection(conn)
	}

	s.CloseAllConn()
}

func (s *Server) handleClientConnection(conn net.Conn) {
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
			continue
		}

		message := string(buffer[:bytesRead])

		if message == "exit" || message == "quit" {
			fmt.Printf("Closing Conn to %s\n", conn.RemoteAddr().String())
			s.CloseConn(conn)
			break
		}
		s.BroadcastMessage(conn, buffer[:bytesRead])
	}
}

func (s *Server) CloseConn(conn net.Conn) {
	s.clients.lock.Lock()

	for i := uint8(0); i < s.clients.maxClients; i++ {
		if s.clients.clientConnections[i] == nil {
			continue
		}
		if s.clients.clientConnections[i].RemoteAddr() == conn.RemoteAddr() {
			conn.Close()
			s.clients.clientConnections[i] = nil
		}
	}

	s.clients.lock.Unlock()
}

func (s *Server) CloseAllConn() {
	s.clients.lock.Lock()

	for i := uint8(0); i < s.clients.maxClients; i++ {
		if s.clients.clientConnections[i] != nil {
			s.clients.clientConnections[i].Close()
		}
		s.clients.clientConnections[i] = nil
	}

	s.clients.lock.Unlock()
}

func (s *Server) RegisterClient(conn net.Conn) bool {
	s.clients.lock.Lock()

	s.clients.connectedClients += 1
	didRegister := false
	for i := uint8(0); i < s.clients.maxClients; i++ {
		if s.clients.clientConnections[i] == nil {
			s.clients.clientConnections[i] = conn
			didRegister = true
			break
		}
	}

	s.clients.lock.Unlock()
	return didRegister
}

func (s *Server) BroadcastMessage(conn net.Conn, message []byte) {
	fmt.Printf("Broadcasting message from %s\n", conn.RemoteAddr().String())

	s.clients.lock.RLock()
	for i := uint8(0); i < s.clients.maxClients; i++ {
		if s.clients.clientConnections[i] == nil {
			continue
		}
		_, err := s.clients.clientConnections[i].Write(message)
		if err != nil {
			fmt.Println(err)
		}
	}
	s.clients.lock.RUnlock()
}

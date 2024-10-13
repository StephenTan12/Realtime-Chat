package main

import (
	"fmt"
	"io"
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
	clientConnections []ClientData
	lock              sync.RWMutex
}

type ClientData struct {
	name string
	conn net.Conn
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
		clientConnections: make([]ClientData, maxClients),
	}

	fmt.Println("Listening to Connections")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}

		go s.HandleClientConnection(conn)
	}

	s.CloseAllConn()
}

func (s *Server) FormatMessage(name string, message string) string {
	return name + ": " + message
}

func (s *Server) HandleClientConnection(conn net.Conn) {
	name, didRegister := s.RegisterClient(conn)
	if !didRegister {
		conn.Write([]byte("Max number of registrations have been reached!\n"))
		return
	}

	fmt.Printf("Registered Client %s\n", name)

	buffer := make([]byte, 512)

	for {
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				s.CloseConn(conn, name)
				return
			}
			fmt.Println(err)
			continue
		}

		message := string(buffer[:bytesRead])

		if message == "exit" || message == "quit" {
			fmt.Printf("Closing Conn to %s\n", conn.RemoteAddr().String())
			s.CloseConn(conn, name)
			break
		}
		s.BroadcastMessage(conn, s.FormatMessage(name, message))
	}
}

func (s *Server) CloseConn(conn net.Conn, name string) {
	s.clients.lock.Lock()

	for i := uint8(0); i < s.clients.maxClients; i++ {
		if s.clients.clientConnections[i].conn == nil {
			continue
		}
		if s.clients.clientConnections[i].conn.RemoteAddr() == conn.RemoteAddr() {
			fmt.Printf("Closed conn from %s\n", conn.RemoteAddr().String())
			conn.Write([]byte("close\b255"))
			conn.Close()
			s.clients.clientConnections[i].conn = nil
			s.clients.connectedClients -= 1
		}
	}

	s.clients.lock.Unlock()

	message := name + " has disconnected\n"
	s.BroadcastMessage(conn, message)

	fmt.Printf("Number of connected clients is now %d\n", s.clients.connectedClients)
}

func (s *Server) CloseAllConn() {
	s.clients.lock.Lock()

	for i := uint8(0); i < s.clients.maxClients; i++ {
		if s.clients.clientConnections[i].conn != nil {
			s.clients.clientConnections[i].conn.Close()
			s.clients.connectedClients -= 1
		}
		s.clients.clientConnections[i].conn = nil
	}

	s.clients.lock.Unlock()

	fmt.Println("Closed all server connections")
}

func (s *Server) RegisterClient(conn net.Conn) (string, bool) {
	nameBuffer := make([]byte, 40)
	bytesRead, _ := conn.Read(nameBuffer)
	name := string(nameBuffer[:bytesRead-1])

	s.clients.lock.Lock()

	s.clients.connectedClients += 1
	didRegister := false
	for i := uint8(0); i < s.clients.maxClients; i++ {
		if s.clients.clientConnections[i].conn == nil {
			s.clients.clientConnections[i] = ClientData{name: name, conn: conn}
			didRegister = true
			break
		}
	}

	s.clients.lock.Unlock()

	fmt.Printf("Number of connected clients is now %d\n", s.clients.connectedClients)

	if didRegister {
		message := name + " has joined!\n"
		s.BroadcastMessage(conn, message)
	}

	return name, didRegister
}

func (s *Server) BroadcastMessage(conn net.Conn, message string) {
	fmt.Printf("Broadcasting message from %s\n", conn.RemoteAddr().String())

	s.clients.lock.RLock()
	for i := uint8(0); i < s.clients.maxClients; i++ {
		if s.clients.clientConnections[i].conn == nil {
			continue
		}
		if s.clients.clientConnections[i].conn.RemoteAddr() == conn.RemoteAddr() {
			continue
		}

		_, err := s.clients.clientConnections[i].conn.Write([]byte(message))
		if err != nil {
			fmt.Println(err)
		}
	}
	s.clients.lock.RUnlock()
}

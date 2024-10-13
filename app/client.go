package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

type Client struct {
	serverPort uint16
}

func (c *Client) ServerIpAddrStr() string {
	return ServerIpStr + ":" + strconv.FormatUint(uint64(c.serverPort), 10)
}

func (c *Client) InitClient() {
	serverConn, err := net.Dial("tcp", c.ServerIpAddrStr())
	if err != nil {
		fmt.Println("Failed to connect to server!")
		os.Exit(1)
	}
	defer serverConn.Close()

	fmt.Println("Connected to server!")
	closeCh := make(chan bool, 1)
	go c.ReceiveMessages(serverConn, closeCh)
	c.SendMessages(serverConn, closeCh)
}

func (c *Client) SendMessages(conn net.Conn, closeCh chan bool) {
	var message string

	fmt.Print("Enter your name: ")

	reader := bufio.NewReader(os.Stdin)
	message, _ = reader.ReadString('\n')
	_, err := conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Failed to register")
		fmt.Println("Closing connection")
		closeCh <- true
		return
	}

	fmt.Println("Start entering your messages!")

	for {
		message, _ = reader.ReadString('\n')
		_, err = conn.Write([]byte(message))
		if err != nil {
			select {
			case <-closeCh:
				fmt.Println("Server has been closed")
				return
			default:
				if err == io.EOF {
					fmt.Println("Server has been closed")
					return
				}
				fmt.Println(err)
				fmt.Println("Failed to write message")
			}
		}

		if message == "exit" || message == "quit" {
			closeCh <- true
			return
		}
	}
}

func (c *Client) ReceiveMessages(conn net.Conn, closeCh chan bool) {
	buffer := make([]byte, 512)
	for {
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Server has been closed")
				closeCh <- true
				return
			}
			select {
			case <-closeCh:
				fmt.Println("Server has been closed")
				return
			default:
				fmt.Println(err)
				fmt.Println("Failed to read message")
			}
		}
		message := string(buffer[:bytesRead])

		if message == "close\b255" {
			fmt.Println("closed")
			closeCh <- true
			return
		}

		fmt.Print(string(buffer[:bytesRead]))
	}
}

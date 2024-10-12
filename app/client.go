package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

type Client struct {
	name       string
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

	go c.ReceiveMessages(serverConn)
	c.SendMessages(serverConn)
}

func (c *Client) FormatMessage(message string) string {
	return c.name + ": " + message
}

func (c *Client) SendMessages(conn net.Conn) {
	var message string
	fmt.Println("Start Entering your messages!")
	for {
		fmt.Scan(&message)
		_, err := conn.Write([]byte(c.FormatMessage(message)))
		if err != nil {
			fmt.Println("failed to write message")
		}

		if message == "exit" || message == "quit" {
			break
		}
	}
}

func (c *Client) ReceiveMessages(conn net.Conn) {
	buffer := make([]byte, 512)
	for {
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			fmt.Println(err)
			fmt.Println("failed to read message")
			break
		}

		fmt.Println(string(buffer[:bytesRead]))
	}
}

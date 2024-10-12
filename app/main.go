package main

import (
	"os"
)

/*
Graceful shutdown of servers and closing connections
- add protocol to dynamic set name
- set name then start sending messages

when server close, send close\b{some special byte} message to all clients
*/

const ServerIpStr = "127.0.0.1"
const ServerPort = 7007
const MaxClients = 254

func main() {
	args := os.Args[1:]

	if args[0] == "server" {
		server := Server{
			port:      ServerPort,
			ipAddrStr: ServerIpStr,
		}
		server.InitServer(MaxClients)
	} else if args[0] == "client" {
		name := args[1]
		client := Client{
			name:       name,
			serverPort: ServerPort,
		}
		client.InitClient()
	}
}

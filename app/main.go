package main

import (
	"os"
)

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
		client := Client{
			serverPort: ServerPort,
		}
		client.InitClient()
	}
}

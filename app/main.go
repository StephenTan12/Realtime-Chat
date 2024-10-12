package main

const MaxClients = 1

func main() {
	server := Server{
		port:      7007,
		ipAddrStr: "127.0.0.1",
	}

	server.InitServer(MaxClients)
}

package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

var servers = []string{
	"localhost:8080",
	"localhost:8081",
}

var counter = 0

func main() {
	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err.Error())
	}

	defer ln.Close()

	for {
		clientConn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleConnection(clientConn)
	}
}

func handleConnection(clientConn net.Conn) {
	defer clientConn.Close()
	counter++
	serverConnection, err := net.Dial("tcp", servers[counter%len(servers)])
	if err != nil {
		log.Printf("Error connecting to server: %v", err)
		return
	}
	defer serverConnection.Close()
	fmt.Printf("Connected to server: %s\n", servers[counter%len(servers)])

	go func() {
		clientCopy, err := io.Copy(serverConnection, clientConn)
		if err != nil {
			log.Printf("Error copying data from client to server: %v", err)
			return
		}

		fmt.Printf("Copied %d bytes from client to server\n", clientCopy)
	}()

	serverCopy, err := io.Copy(clientConn, serverConnection)
	if err != nil {
		log.Printf("Error copying data from server to client: %v", err)
		return
	}
	fmt.Printf("Copied %d bytes from server to client\n", serverCopy)
}

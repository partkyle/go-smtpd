package main

import (
	"log"
	"net"
	"os"
)

var logger = log.New(os.Stdout, "[smtpd] ", log.Lshortfile|log.LstdFlags)

func handleConnection(conn net.Conn) {
	conv := &Conversation{
		conn: conn,
	}

	conv.Run()
}

func main() {
	listener, err := net.Listen("tcp", ":2525")
	if err != nil {
		log.Fatalf("Error listening: %q", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Error accepting:", err)
		}

		go handleConnection(conn)
	}
}

package main

import (
	"log"
	"net"
	"os"
	"time"
)

var logger = log.New(os.Stdout, "[smtpd] ", log.Lshortfile|log.LstdFlags)

func handleConnection(conn net.Conn) {
	conv := &Conversation{
		conn:        conn,
		idleTimeout: 5 * time.Second,
	}

	conv.Run()
}

func main() {
	listener, err := net.Listen("tcp", ":2525")
	if err != nil {
		log.Fatalf("Error listening: %q", err)
	}

	log.Printf("listening for SMTP on %s", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Error accepting:", err)
		}

		go handleConnection(conn)
	}
}

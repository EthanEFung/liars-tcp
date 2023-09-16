package main

import (
	"log"
	"net"
)

func main() {
	s := NewServer()
	go s.Serve()

	listener, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatalf("cannot listen on port 8888 ::: %s", err.Error())
	}
	defer listener.Close()
	log.Println("listening on port 8888")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("cannot accept connection ::: %s\n", err.Error())
			continue
		}
		log.Printf("connecting client %s\n", conn.RemoteAddr().String())
		go s.ConnectClient(conn)
	}
}

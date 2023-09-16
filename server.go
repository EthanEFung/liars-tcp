package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Server interface {
	Serve()
	ConnectClient(conn net.Conn)
}

func NewServer() Server {
	return &server{
		rooms: make(map[string]Room),
		cmds:  make(chan Command),
	}
}

type server struct {
	rooms map[string]Room
	cmds  chan Command
}

func (s *server) Serve() {
	for cmd := range s.cmds {
		switch cmd.ID() {
		case CMD_NAME:
			s.name(cmd)
		case CMD_JOIN:
			s.join(cmd)
		case CMD_ROOMS:
			s.listRooms(cmd)
		case CMD_QUIT:
			s.quit(cmd)
		}
	}
}

func (s *server) ConnectClient(conn net.Conn) {
	c := NewClient(conn, s.cmds)

	log.Printf("connected client %s\n", conn.RemoteAddr().String())
	c.Println("welcome to liars dice!")

	c.ReadInput()
}

func (s *server) name(cmd Command) {
	c := cmd.Client()
	args := cmd.Args()
	prev := c.Name()
	c.SetName(args[1])
	c.Printf("set name to %s\n", c.Name())

	if r := c.Room(); r != nil {
		r.Broadcast(c, fmt.Sprintf("%s changed name to %s", prev, c.Name()))
	}
}

func (s *server) join(cmd Command) {
	args := cmd.Args()
	c := cmd.Client()
	roomname := args[1]
	r, exists := s.rooms[roomname]
	if !exists {
		r = NewRoom(roomname)
		s.rooms[roomname] = r
	}
	r.AddMember(c)

	curr := c.Room()
	if curr != nil {
		curr.BootMember(c)
	}

	c.SetRoom(r)

	r.Broadcast(c, fmt.Sprintf("%s has joined the room\n", c.Name()))
	c.Printf("welcome to room: %s\n", r.Name())
}

func (s *server) listRooms(cmd Command) {
	c := cmd.Client()
	names := []string{}
	for name := range s.rooms {
		names = append(names, name)
	}
	c.Printf("available rooms: %s\n", strings.Join(names, " "))
}

func (s *server) quit(cmd Command) {
	c := cmd.Client()
	r := c.Room()
	log.Printf("disconnected client %s\n", c.Addr().String())
	if r != nil {
		r.BootMember(c)
	}
	c.Println("thanks for stopping by!")
	c.Close()
}

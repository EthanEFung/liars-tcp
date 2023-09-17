package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

var ErrNotInRoom = errors.New("must be in a room")
var ErrInternal = errors.New("internal error")

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
		case CMD_PLAY:
			s.playGame(cmd)
		case CMD_START_GAME:
			s.startGame(cmd)
		case CMD_DICE:
			s.dice(cmd)
		case CMD_WAGER:
			s.wager(cmd)
		case CMD_LIAR:
			s.liar(cmd)
		case CMD_RESET_GAME:
			s.resetGame(cmd)
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

	r.Broadcast(c, fmt.Sprintf("%s has joined the room", c.Name()))
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

func (s *server) playGame(cmd Command) {
	c := cmd.Client()
	r := c.Room()
	g := r.Game()
	if r == nil {
		c.Error(ErrNotInRoom)
		return
	}
	if g == nil {
		// this should be impossible
		c.Error(ErrInternal)
		return
	}
	p, err := g.AddPlayer(c.Addr())
	if err != nil {
		c.Error(err)
		return
	} 
	c.SetPlayer(p)
	r.Broadcast(c, c.Name() + " joined the game")
	c.Println("great! you have joined the game")
}

func (s *server) startGame(cmd Command) {
	c := cmd.Client()
	r := c.Room()
	g := r.Game()
	if r == nil {
		c.Error(ErrNotInRoom)
		return
	} 
	if err := g.Start(); err != nil {
		c.Error(err)
		return
	}
	r.Broadcast(c, "game has started enter '/dice' to see what dice you've been given")
	c.Println("game has started enter '/dice' to see what dice you've been given")
}

func (s *server) dice(cmd Command) {
	c := cmd.Client()
	p := c.Player()
	dice := p.Dice()
	b := strings.Builder{}
	b.WriteString("you've rolled:\n")
	for x, count := range dice {
		b.WriteString(fmt.Sprintf("%d %d's\n", count, x))
	}
	c.Printf(b.String())
}

func (s *server) wager(cmd Command) {
	c := cmd.Client()
	r := c.Room()
	g := r.Game()
	p := c.Player()
	if r == nil {
		c.Error(ErrNotInRoom)
	}
	if g == nil {
		// this should be impossible
		c.Error(ErrInternal)
		return
	}
	if p == nil {
		c.Error(ErrInternal)
		return
	}
	args := cmd.Args()
	if len(args) < 3 {
		c.Error(errors.New("'/wager' command requires `[count] [face]`"))
		return
	}
	count, err := strconv.Atoi(args[1])
	if err != nil {
		c.Error(errors.New("please provide a numeric value for count"))
		return
	}
	face, err := strconv.Atoi(args[2])
	if err != nil {
		c.Error(errors.New("please provide a numeric value for faces"))
		return
	}
	w, err := NewWager(face, count)
	if err != nil  {
		c.Error(err)
		return
	}
	if err := g.SetWager(p, w); err != nil  {
		c.Error(err)
		return
	}
	w, err = g.Wager()
	if err != nil {
		c.Error(err)
		return
	}
	c.Printf("you've wagered %d %d's\n", w.Count(), w.Face())
	r.Broadcast(c, fmt.Sprintf("%s wagers %d %d's", c.Name(), w.Count(), w.Face()))
}

func (s *server) liar(cmd Command) {
	c := cmd.Client()
	p := c.Player()
	r := c.Room()
	g := r.Game()
	if r == nil {
		c.Error(ErrNotInRoom)
	}
	if g == nil {
		// this should be impossible
		c.Error(ErrInternal)
		return
	}
	if p == nil {
		c.Error(ErrInternal)
		return
	}
	loser, err := g.Call(p)
	if err != nil {
		c.Error(err)
		return
	}
	if p == loser {
		c.Println("you lost!")
		r.Broadcast(c ,fmt.Sprintf("%s lost the round!", c.Name()))
	} else {
		c.Println("you won the round!")
		r.Broadcast(c, fmt.Sprintf("%s won the round!", c.Name()))
	}
}

func (s *server) resetGame(cmd Command) {
	c := cmd.Client()
	r := c.Room()
	g := r.Game()
	if r == nil {
		c.Error(ErrNotInRoom)
	}
	if g == nil {
		// this should be impossible
		c.Error(ErrInternal)
		return
	}
	// ehh...for now just allow the game to be reset at any point
	err := r.ResetGame()
	if err != nil {
		c.Error(err)
		return
	}
	c.Println("game has been reset")
	r.Broadcast(c, "game has been reset")
}

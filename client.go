package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type Client interface {
	// process commands for the server run time
	ReadInput()
	Close() error

	// how to message the client
	Error(error)
	Println(string)
	Printf(str string, args ...any)

	// public attributes to share with the rest of the system
	Addr() net.Addr
	Name() string
	SetName(string)
	Room() Room
	SetRoom(Room)
	Player() Player
	SetPlayer(Player)
}

func NewClient(conn net.Conn, cmds chan Command) Client {
	return &client{
		conn: conn,
		name: "anonymous",
		cmds: cmds,
	}
}

type client struct {
	conn   net.Conn
	name   string
	cmds   chan<- Command
	room   Room
	player Player
}

func (c *client) ReadInput() {
	for {
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				c.cmds <- NewCommand(CMD_QUIT, c, []string{"/quit"})
			} else if errors.Is(err, net.ErrClosed) {
				// we'll just ignore
			} else {
				log.Printf("unexpected error from client %s ::: %s\n", c.Addr(), err.Error())
			}
			return
		}
		msg = strings.Trim(msg, "\r\n")

		args := strings.Split(msg, " ")
		cmd := strings.TrimSpace(args[0])

		switch cmd {
		case "/name":
			c.cmds <- NewCommand(CMD_NAME, c, args)
		case "/join":
			c.cmds <- NewCommand(CMD_JOIN, c, args)
		case "/rooms":
			c.cmds <- NewCommand(CMD_ROOMS, c, args)
		case "/quit":
			c.cmds <- NewCommand(CMD_QUIT, c, args)
		case "/play":
			c.cmds <- NewCommand(CMD_PLAY, c, args)
		case "/start":
			c.cmds <- NewCommand(CMD_START_GAME, c, args)
		case "/dice":
			c.cmds <- NewCommand(CMD_DICE, c, args)
		case "/bet":
			c.cmds <- NewCommand(CMD_WAGER, c, args)
		case "/liar":
			c.cmds <- NewCommand(CMD_LIAR, c, args)
		case "/reset":
			c.cmds <- NewCommand(CMD_RESET_GAME, c, args)
		default:
			c.Error(fmt.Errorf("command '%s' is not recognized", cmd))
		}
	}
}

func (c *client) Error(err error) {
	c.conn.Write([]byte("! " + err.Error() + "\n"))
}

func (c *client) Println(str string) {
	c.conn.Write([]byte("> " + str + "\n"))
}

func (c *client) Printf(str string, args ...any) {
	c.conn.Write([]byte(fmt.Sprintf("> "+str, args...)))
}

func (c *client) Addr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *client) Name() string {
	return c.name
}

func (c *client) SetName(name string) {
	c.name = name
}

func (c *client) Room() Room {
	return c.room
}

func (c *client) SetRoom(r Room) {
	c.room = r
}

func (c *client) Close() error {
	return c.conn.Close()
}

func (c *client) Player() Player {
	return c.player
}

func (c *client) SetPlayer(p Player) {
	c.player = p
}

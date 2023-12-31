package main

type CommandID int

const (
	CMD_NAME CommandID = iota
	CMD_JOIN
	CMD_ROOMS
	CMD_QUIT

	CMD_PLAY
	CMD_START_GAME
	CMD_DICE
	CMD_WAGER
	CMD_LIAR
	CMD_RESET_GAME
)

type Command interface {
	ID() CommandID
	Client() Client
	Args() []string
}

func NewCommand(id CommandID, c Client, args []string) Command {
	return &command{
		id:   id,
		c:    c,
		args: args,
	}
}

type command struct {
	id   CommandID
	c    Client
	args []string
}

func (c *command) ID() CommandID {
	return c.id
}

func (c *command) Client() Client {
	return c.c
}

func (c *command) Args() []string {
	return c.args
}

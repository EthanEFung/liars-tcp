package main

import (
	"errors"
	"net"
)

type GameStateType int

const (
	GAME_STATE_PENDING GameStateType = iota
	GAME_STATE_PLAYING
	GAME_STATE_PLAYED
)

var ErrGameNotStarted = errors.New("game has not started")
var ErrGameInProgress = errors.New("game is in progress")
var ErrGamePlayed = errors.New("game has finished")

type Game interface {
	GameState
	SetState(GameState)

	start() error
	setWager(Player, Wager) error
	addPlayer(Player) error
	removePlayer(Player) error
	allDice() map[int]int
	winner() (Player, error)
	startNewRound() error
}

func NewGame() Game {
	g := &game{
		players: []Player{},
		dice:    make(map[int]int),
	}
	g.state = &gameStatePending{g}
	return g
}

type game struct {
	state   GameState
	wager   Wager
	idx     int
	player  Player
	players []Player
	dice    map[int]int
}

func (g *game) SetState(state GameState) {
	g.state = state
}
func (g *game) Type() GameStateType {
	return g.state.Type()
}
func (g *game) Player() Player {
	return g.players[g.idx]
}
func (g *game) Players() []Player {
	return g.players
}
func (g *game) AddPlayer(addr net.Addr) (Player, error) {
	return g.state.AddPlayer(addr)
}
func (g *game) Start() error {
	return g.state.Start()
}
func (g *game) Wager() (Wager, error) {
	return g.wager, nil
}
func (g *game) SetWager(p Player, w Wager) error {
	return g.state.SetWager(p, w)
}
func (g *game) Call(p Player) (Player, error) {
	return g.state.Call(p)
}
func (g *game) Winner() (Player, error) {
	return g.state.Winner()
}
func (g *game) start() error {
	// maybe it's fine that the user has already rolled their dice?
	g.player = g.players[0]
	return nil
}
func (g *game) setWager(p Player, w Wager) error {
	g.wager = w
	g.idx++
	if g.idx >= len(g.players) {
		g.idx = 0
	}
	return nil
}
func (g *game) allDice() map[int]int {
	d := make(map[int]int)
	for _, p := range g.Players() {
		for x, count := range p.Dice() {
			d[x] += count
		}
	}
	return d
}
func (g *game) addPlayer(p Player) error {
	g.players = append(g.players, p)
	return nil
}
func (g *game) removePlayer(p Player) error {
	prev := len(g.players)
	for i, player := range g.players {
		if p == player {
			g.players = append(g.players[:i], g.players[i+1:]...)
			g.idx = i
			if g.idx >= len(g.players) {
				g.idx = 0
			}
		}
	}
	if len(g.players) != prev-1 {
		return errors.New("unsuccessfully removed player")
	}
	return nil
}
func (g *game) winner() (Player, error) {
	if len(g.players) != 1 {
		return nil, errors.New("winner function was called but winner is undetermined")
	}
	return g.players[0], nil
}
func (g *game) startNewRound() error {
	g.wager = nil
	for _, p := range g.players {
		p.Roll()
	}
	return nil
}

type GameState interface {
	Type() GameStateType
	Player() Player
	Players() []Player
	AddPlayer(net.Addr) (Player, error)
	Wager() (Wager, error)
	SetWager(Player, Wager) error
	Call(Player) (Player, error)
	Winner() (Player, error)
	Start() error
}

type gameStatePending struct {
	Game
}

func (gsp *gameStatePending) Type() GameStateType {
	return GAME_STATE_PENDING
}
func (gsp *gameStatePending) Players() []Player {
	return gsp.Game.Players()
}
func (gsp *gameStatePending) AddPlayer(addr net.Addr) (Player, error) {
	for _, player := range gsp.Game.Players() {
		if player.Addr() == addr {
			return nil, errors.New("player is already playing")
		}
	}
	p := NewPlayer(addr)
	if err := gsp.Game.addPlayer(p); err != nil {
		return nil, err
	}
	return p, nil
}
func (gsp *gameStatePending) Start() error {
	g := gsp.Game
	if len(g.Players()) < 2 {
		return errors.New("cannot start a game without at least two players")
	}
	g.SetState(&gameStatePlaying{g})
	return g.start()
}
func (gsp *gameStatePending) Wager() (Wager, error) {
	return nil, ErrGameNotStarted
}
func (gsp *gameStatePending) SetWager(Player, Wager) error {
	return ErrGameNotStarted
}
func (gsp *gameStatePending) Call(Player) (Player, error) {
	return nil, ErrGameNotStarted
}
func (gsp *gameStatePending) Winner() (Player, error) {
	return nil, ErrGameNotStarted
}

type gameStatePlaying struct {
	Game
}

func (gsp *gameStatePlaying) Type() GameStateType {
	return GAME_STATE_PLAYING
}
func (gsp *gameStatePlaying) Players() []Player {
	return gsp.Game.Players()
}
func (gsp *gameStatePlaying) AddPlayer(addr net.Addr) (Player, error) {
	return nil, ErrGameInProgress
}
func (gsp *gameStatePlaying) Start() error {
	return ErrGameInProgress
}
func (gsp *gameStatePlaying) Wager() (Wager, error) {
	return gsp.Game.Wager()
}
func (gsp *gameStatePlaying) SetWager(p Player, w Wager) error {
	g := gsp.Game
	if g.Player() != p {
		return errors.New("only current player can set wager")
	}
	curr, err := g.Wager()
	if err != nil {
		return err
	}
	if curr != nil && w.LTE(curr) {
		return errors.New("wager must be greater than current")
	}
	g.setWager(p, w)
	return nil
}

/*
Call returns the player that lost the wager
*/
func (gsp *gameStatePlaying) Call(p Player) (Player, error) {
	g := gsp.Game
	w, err := g.Wager()
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, errors.New("no wager has been made")
	}
	if g.Player() != p {
		return nil, errors.New("only current player can call")
	}
	var loser Player
	dice := g.allDice()
	total := dice[w.Face()]
	players := g.Players()

	if total < w.Count() {
		for i, player := range players {
			if p == player {
				if i == 0 {
					loser = players[len(players)-1]
				} else {
					loser = players[i-1]
				}
				break
			}
		}
		if loser == nil {
			return nil, errors.New("could not find losing player")
		}
	} else {
		if p == nil {
			return nil, errors.New("could not find losing player")
		}
		loser = p
	}
	g.removePlayer(loser)

	// the total is greater or equal to the wager so the player who calls loses
	winner, err := g.Winner()
	if err != nil && !errors.Is(err, ErrGameInProgress) {
		return nil, err
	}
	if winner != nil {
		g.SetState(&gameStatePlayed{g})
	}

	g.startNewRound()
	return loser, nil
}
func (gsp *gameStatePlaying) Winner() (Player, error) {
	if len(gsp.Game.Players()) != 1 {
		return nil, ErrGameInProgress
	}
	return gsp.Game.winner()
}

type gameStatePlayed struct {
	Game
}

func (gsp *gameStatePlayed) Type() GameStateType {
	return GAME_STATE_PLAYED
}
func (gsp *gameStatePlayed) Players() []Player {
	return gsp.Game.Players()
}
func (gsp *gameStatePlayed) AddPlayer(net.Addr) (Player, error) {
	return nil, ErrGamePlayed
}
func (gsp *gameStatePlayed) Wager() (Wager, error) {
	return nil, ErrGamePlayed
}
func (gsp *gameStatePlayed) SetWager(Player, Wager) error {
	return ErrGamePlayed
}
func (gsp *gameStatePlayed) Call(Player) (Player, error) {
	return nil, ErrGamePlayed
}
func (gsp *gameStatePlayed) Winner() (Player, error) {
	return gsp.Game.winner()
}
func (gsp *gameStatePlayed) Start() error {
	return ErrGamePlayed
}

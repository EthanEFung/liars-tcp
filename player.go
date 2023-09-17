package main

import (
	"math/rand"
	"net"
	"time"
)

type Player interface {
	Addr() net.Addr
	Dice() map[int]int 
	Roll()
}

func NewPlayer(addr net.Addr) Player {
	d := make(map[int]int, 5)
	r := rand.New(rand.NewSource(time.Now().UnixMicro()))
	for i := 1; i <= 5; i++ {
		x := r.Intn(6)
		d[x+1]++
	}

	return &player{
		addr: addr,
		dice: d,
	}
}

type player struct {
	addr net.Addr
	dice map[int]int
}

func (p *player) Dice() map[int]int {
	return p.dice
}

func (p *player) Addr() net.Addr {
	return p.addr
}

func (p *player) Roll() {
	d := make(map[int]int, 5)
	r := rand.New(rand.NewSource(time.Now().UnixMicro()))
	for i := 1; i <= 5; i++ {
		x := r.Intn(6)
		d[x+1]++
	}
	p.dice = d
}

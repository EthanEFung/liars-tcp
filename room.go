package main

import (
	"fmt"
	"net"
)

type Room interface {
	Name() string
	Broadcast(sender Client, msg string)
	AddMember(client Client)
	BootMember(member Client)
}

func NewRoom(name string) Room {
	return &room{
		name:    name,
		members: make(map[net.Addr]Client),
	}
}

type room struct {
	name    string
	members map[net.Addr]Client
}

func (r *room) Name() string {
	return r.name
}

func (r *room) Broadcast(sender Client, msg string) {
	for addr, m := range r.members {
		if addr != sender.Addr() {
			m.Println(msg)
		}
	}
}

func (r *room) AddMember(client Client) {
	r.members[client.Addr()] = client
}

func (r *room) BootMember(member Client) {
	if member.Room() != nil {
		delete(r.members, member.Addr())
		r.Broadcast(member, fmt.Sprintf("%s has left the room", member.Name()))
	}
}

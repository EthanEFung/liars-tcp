package main

import "errors"

type Wager interface {
	Face() int
	Count() int
	LTE(Wager) bool
}

func NewWager(face, count int) (Wager, error) {
	if (face < 0 || face > 6) {
		return nil, errors.New("invalid wager (only 6 faces on a die)")
	}
	return &wager{
		face,
		count,
	}, nil
}

type wager struct {
	face int
	count int
}

func(w *wager) LTE(other Wager) bool {
	f, c := w.face, w.count
	if c < other.Count() {
		return true
	}
	if c == other.Count() && f <= other.Face() {
		return true
	}
	return false
}

func (w *wager) Face() int {
	return w.face
} 

func (w *wager) Count() int {
	return w.count
}

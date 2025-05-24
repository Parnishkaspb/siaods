// Реализация CRDT Logoot для редактирования текста (упрощённая)
// Реализует вставку символов в позицию с уникальным идентификатором

package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type PositionComponent struct {
	Digit  int
	SiteID string
}

type Position []PositionComponent

func (p Position) Less(other Position) bool {
	for i := 0; i < len(p) && i < len(other); i++ {
		if p[i].Digit != other[i].Digit {
			return p[i].Digit < other[i].Digit
		}
		if p[i].SiteID != other[i].SiteID {
			return p[i].SiteID < other[i].SiteID
		}
	}
	return len(p) < len(other)
}

func (p Position) Equal(other Position) bool {
	if len(p) != len(other) {
		return false
	}
	for i := range p {
		if p[i].Digit != other[i].Digit || p[i].SiteID != other[i].SiteID {
			return false
		}
	}
	return true
}

type Atom struct {
	Position Position
	Char     rune
	Deleted  bool
}

// Protocol message
type Message struct {
	From      string
	Operation string // "insert" or "delete"
	Atom      Atom
}

// Реплика документа
type Replica struct {
	SiteID string
	Atoms  []Atom
	Inbox  chan Message
	Outbox chan Message
	Clock  int
	mu     sync.Mutex
}

func NewReplica(id string) *Replica {
	return &Replica{
		SiteID: id,
		Inbox:  make(chan Message, 100),
		Outbox: make(chan Message, 100),
	}
}

func (r *Replica) findInsertIndex(pos Position) int {
	for i, a := range r.Atoms {
		if pos.Less(a.Position) {
			return i
		}
	}
	return len(r.Atoms)
}

func (r *Replica) localInsert(index int, ch rune) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pos := r.generatePosition(index)
	atom := Atom{Position: pos, Char: ch}
	insertIdx := r.findInsertIndex(pos)
	r.Atoms = append(r.Atoms, Atom{})
	copy(r.Atoms[insertIdx+1:], r.Atoms[insertIdx:])
	r.Atoms[insertIdx] = atom

	r.Outbox <- Message{From: r.SiteID, Operation: "insert", Atom: atom}
}

func (r *Replica) generatePosition(index int) Position {
	var left, right Position
	if index > 0 {
		left = r.Atoms[index-1].Position
	} else {
		left = Position{{Digit: 0, SiteID: ""}}
	}
	if index < len(r.Atoms) {
		right = r.Atoms[index].Position
	} else {
		right = Position{{Digit: 32767, SiteID: ""}}
	}

	// Найти позицию между left и right (простой случай, 1 уровень)
	newDigit := (left[0].Digit + right[0].Digit) / 2
	if newDigit == left[0].Digit {
		newDigit++ // избегаем коллизий
	}

	return Position{{Digit: newDigit, SiteID: r.SiteID}}
}

func (r *Replica) apply(msg Message) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch msg.Operation {
	case "insert":
		insertIdx := r.findInsertIndex(msg.Atom.Position)
		r.Atoms = append(r.Atoms, Atom{})
		copy(r.Atoms[insertIdx+1:], r.Atoms[insertIdx:])
		r.Atoms[insertIdx] = msg.Atom
	case "delete":
		for i := range r.Atoms {
			if r.Atoms[i].Position.Equal(msg.Atom.Position) {
				r.Atoms[i].Deleted = true
				break
			}
		}
	}
}

func (r *Replica) Text() string {
	var sb strings.Builder
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, a := range r.Atoms {
		if !a.Deleted {
			sb.WriteRune(a.Char)
		}
	}
	return sb.String()
}

func (r *Replica) RunProtocol(peers []*Replica) {
	go func() {
		for msg := range r.Outbox {
			for _, peer := range peers {
				if peer.SiteID != r.SiteID {
					peer.Inbox <- msg
				}
			}
		}
	}()

	go func() {
		for msg := range r.Inbox {
			r.apply(msg)
		}
	}()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	r1 := NewReplica("A")
	r2 := NewReplica("B")
	r3 := NewReplica("C")

	replicas := []*Replica{r1, r2, r3}
	for _, r := range replicas {
		r.RunProtocol(replicas)
	}

	// Тест 1: одна реплика вставляет текст
	for _, ch := range "hello" {
		r1.localInsert(len(r1.Text()), ch)
	}
	time.Sleep(500 * time.Millisecond)

	// Тест 2: вторая реплика удаляет символ и вставляет другой
	// (предположим что удаляем второй символ и вставляем 'a')
	if len(r2.Atoms) > 1 {
		atom := r2.Atoms[1]
		r2.Outbox <- Message{From: r2.SiteID, Operation: "delete", Atom: atom}
		r2.localInsert(1, 'a')
	}

	time.Sleep(1 * time.Second)
	fmt.Println("r1:", r1.Text())
	fmt.Println("r2:", r2.Text())
	fmt.Println("r3:", r3.Text())
}

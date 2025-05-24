package crdt

import (
	"fmt"
	"strings"
)

// PositionComponent — одна часть позиции (логическая координата символа)
type PositionComponent struct {
	Digit  int    // числовое значение позиции
	SiteID string // идентификатор реплики (для разрешения конфликтов)
}

// Position — позиция символа в тексте, может быть вложенной
type Position []PositionComponent

// Less — сравнение позиций для упорядочивания символов
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

// Equal — сравнение позиций на равенство
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

// Atom — представление символа в CRDT, включая позицию и статус удаления
type Atom struct {
	Position Position // уникальная логическая позиция
	Char     rune     // сам символ
	Deleted  bool     // помечен ли как удалённый
}

// Message — сетевое сообщение между репликами
type Message struct {
	From      string // отправитель
	Operation string // тип операции: "insert" или "delete"
	Atom      Atom   // содержимое операции
}

// command — внутренние команды для управления репликой (вставка, получение текста и т.п.)
type command struct {
	kind string
	data any
	resp chan any // канал для получения ответа, если он нужен
}

// Replica — одна реплика документа с CRDT-логикой
type Replica struct {
	SiteID string       // уникальный ID этой реплики
	Atoms  []Atom       // список всех атомов (символов)
	Inbox  chan Message // входящие сетевые сообщения от других реплик
	Outbox chan Message // исходящие сообщения, которые нужно переслать другим
	cmds   chan command // локальные команды (например, вставить символ или вернуть текст)
}

func NewReplica(id string) *Replica {
	r := &Replica{
		SiteID: id,
		Inbox:  make(chan Message, 100),
		Outbox: make(chan Message, 100),
		cmds:   make(chan command, 100),
	}
	go r.run()
	return r
}

// run — основной цикл обработки команды/сообщений внутри реплики
func (r *Replica) run() {
	for {
		select {
		case msg := <-r.Inbox: // получена операция с другой реплики
			r.apply(msg)

		case cmd := <-r.cmds: // локальная команда
			switch cmd.kind {
			case "insert":
				// вставка нового символа
				req := cmd.data.(struct {
					index int
					ch    rune
				})
				pos := r.generatePosition(req.index)
				atom := Atom{Position: pos, Char: req.ch}
				idx := r.findInsertIndex(pos)
				// вставка в локальный список атомов
				r.Atoms = append(r.Atoms, Atom{}) // расширение среза
				copy(r.Atoms[idx+1:], r.Atoms[idx:])
				r.Atoms[idx] = atom
				// печать и отправка другим репликам
				fmt.Printf("[INSERT] %s: '%c' at index %d -> position %+v\n", r.SiteID, req.ch, req.index, pos)
				r.Outbox <- Message{From: r.SiteID, Operation: "insert", Atom: atom}

			case "text":
				// формируем итоговый текст (без удалённых символов)
				var sb strings.Builder
				for _, a := range r.Atoms {
					if !a.Deleted {
						sb.WriteRune(a.Char)
					}
				}
				cmd.resp <- sb.String()
			}
		}
	}
}

// apply — применяет входящее сетевое сообщение (insert/delete)
func (r *Replica) apply(msg Message) {
	fmt.Printf("[APPLY] %s received %s: '%c' at %+v\n", r.SiteID, msg.Operation, msg.Atom.Char, msg.Atom.Position)
	switch msg.Operation {
	case "insert":
		idx := r.findInsertIndex(msg.Atom.Position)
		r.Atoms = append(r.Atoms, Atom{})
		copy(r.Atoms[idx+1:], r.Atoms[idx:])
		r.Atoms[idx] = msg.Atom

	case "delete":
		for i := range r.Atoms {
			if r.Atoms[i].Position.Equal(msg.Atom.Position) {
				r.Atoms[i].Deleted = true
				break
			}
		}
	}
}

// generatePosition — генерирует логическую позицию между двумя соседними атомами
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
	newDigit := (left[0].Digit + right[0].Digit) / 2
	if newDigit == left[0].Digit {
		newDigit++ // примитивная защита от коллизий, но не поддерживает вложенность
	}
	return Position{{Digit: newDigit, SiteID: r.SiteID}}
}

// findInsertIndex — определяет индекс, куда вставлять атом с заданной позицией
func (r *Replica) findInsertIndex(pos Position) int {
	for i, a := range r.Atoms {
		if pos.Less(a.Position) {
			return i
		}
	}
	return len(r.Atoms)
}

// LocalInsert — запрос на вставку символа в позицию index
func (r *Replica) LocalInsert(index int, ch rune) {
	r.cmds <- command{kind: "insert", data: struct {
		index int
		ch    rune
	}{index, ch}}
}

// Text — получает текущий текст (без удалённых символов)
func (r *Replica) Text() string {
	resp := make(chan any)
	r.cmds <- command{kind: "text", resp: resp}
	return (<-resp).(string)
}

// RunProtocol — пересылает все сообщения из Outbox всем другим репликам
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
}

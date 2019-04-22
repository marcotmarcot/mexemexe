package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	m := newMexemexe()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text()
		if len(command) == 0 {
			continue
		}
		switch command[0] {
		case 'q':
			return
		case 't':
			// Adds the following card to the table.
			c := parseCard(command)
			if c == 0 {
				fmt.Println("Invalid command: ", command)
				continue
			}
			m.t.newCard(c)
		case 'h':
			// Adds the following card to the hand.
			c := parseCard(command)
			if c == 0 {
				fmt.Println("Invalid command: ", command)
				continue
			}
			m.h.newCard(c)
			// log.Println(m.h.cs)
		case 'c':
			// Checks if it's possible to play any card on the hand.
			m.findCard()
		case 'p':
			c := parseCard(command)
			if c == 0 {
				fmt.Println("Invalid command: ", command)
				continue
			}
			m.h.removeCard(c)
			m.t.newCard(c)
		case 'r':
			c := parseCard(command)
			if c == 0 {
				fmt.Println("Invalid command: ", command)
				continue
			}
			m.t.removeCard(c)
		default:
			fmt.Println("Invalid command: ", command)
		}
	}
}

type mexemexe struct {
	t *table
	h *table
}

func newMexemexe() *mexemexe {
	return &mexemexe{newTable(), newTable()}
}

func (m *mexemexe) findCard() {
	log.Println("findCard hand", m.h.cs)
	log.Println("findCard table", m.t.cs)
	for _, c := range m.h.cs {
		log.Println("findCard", c)
		if m.t.check(c) {
			fmt.Println(c)
			return
		}
	}
	for i := range m.h.cs {
		if g := m.h.findGame(kindGame, i, true); g != nil {
			fmt.Println(g)
			return
		}
		if g := m.h.findGame(seqGame, i, true); g != nil {
			fmt.Println(g)
			return
		}
	}
	fmt.Println("Not found")
}

type table struct {
	cs []card
	s map[card]*states
}

func newTable() *table {
	s := map[card]*states{}
	for c := 1; c < 14 * 4; c++ {
		s[card(c)] = newStates()
	}
	return &table{nil, s}
}

func (t *table) newCard(c card) {
	insert(&t.cs, c)
	t.s[c].update(noState, inTable)
}

func (t *table) removeCard(c card) {
	for i, tc := range t.cs {
		if tc == c {
			t.cs = append(t.cs[:i], t.cs[i+1:]...)
			break
		}
	}
	t.s[c].update(inTable, noState)
}

func (t *table) check(c card) bool {
	t.newCard(c)
	defer t.removeCard(c)
	// defer log.Println(t.s)
	return newProcessing(t).check(0)
}

func (t *table) findGame(ty gameType, i int, stopAtThree bool) []card {
	var g []card
	c := t.cs[i]
	log.Println("findGame", c)
	cn := c.number()
	if ty == kindGame {
		for s := 0; s < 4; s++ {
			nc := newCard(cn, suit(s))
			if !t.s[nc].has(inTable) {
				continue
			}
			g = append(g, nc)
			if stopAtThree && len(g) == 3 {
				return g
			}
		}
	} else {
		cs := c.suit()
		g = append(g, c)
		for n := cn + 1; n <= 14; n++ {
			en := n
			if en == 14 {
				en = 1
			}
			nc := newCard(en, cs)
			if !t.s[nc].has(inTable) {
				break
			}
			g = append(g, nc)
			if stopAtThree && len(g) == 3 {
				return g
			}
		}
	}
	if len(g) < 3 {
		return nil
	}
	return g
}

type states struct {
	ss []state
}

func newStates() *states {
	return &states{[]state{noState, noState}}
}

func (ss *states) String() string {
	return fmt.Sprintf("%v", ss.ss)
}

func (ss *states) has(s state) bool {
	for _, st := range ss.ss {
		if st == s {
			return true
		}
	}
	return false
}

func (ss *states) update(from, to state) {
	// log.Println("states", ss)
	for i, s := range ss.ss {
		if s == from {
			ss.ss[i] = to
			return
		}
	}
	log.Fatal("No from found ", from, to)
}

type state int

const (
	noState state = iota
	inTable
	inProcessing
)

type processing struct {
	gs []*game
	t *table
}

func newProcessing(t *table) *processing {
	return &processing{nil, t}
}

func (p *processing) check(i int) bool {
	if i == len(p.t.cs) {
		for _, c := range p.t.cs {
			if p.t.s[c].has(inTable) {
				return false
			}
		}
		return true
	}
	c := p.t.cs[i]
	log.Println("p.gs", p.gs)
	log.Println("check", i, len(p.t.cs), c)
	if !p.t.s[c].has(inTable) {
		log.Println("InTable")
		return p.check(i + 1)
	}
	if p.buildGame(kindGame, i) {
		return true
	}
	if p.buildGame(seqGame, i) {
		return true
	}
	if c.number() == 1 {
		return p.check(i + 1)
	}
	return false
}

func (p *processing) buildGame(ty gameType, i int) bool {
	g := p.t.findGame(ty, i, false)
	if g == nil {
		return false
	}
	return p.addGame(i, g)
}

func (p *processing) addGame(i int, g []card) bool {
	p.gs = append(p.gs, newGame(p.t, g))
	defer p.destroyGame()
	if p.check(i + 1) {
		return true
	}
	for p.dropGame() {
		if p.check(i + 1) {
			return true
		}
	}
	return false
}

func (p *processing) destroyGame() {
	log.Println("destroy", len(p.gs))
	p.gs[len(p.gs)-1].destroy()
	p.gs = p.gs[:len(p.gs)-1]
}

func (p *processing) dropGame() bool {
	return p.gs[len(p.gs)-1].dropGame()
}

type game struct {
	t *table
	cs []card
}

func newGame(t *table, cs []card) *game {
	for _, c := range cs {
		t.s[c].update(inTable, inProcessing)
	}
	return &game{t, cs}
}

func (g *game) String() string {
	return fmt.Sprintf("%v", g.cs)
}

func (g *game) destroy() {
	for _, c := range g.cs {
		g.t.s[c].update(inProcessing, inTable)
	}
}

func (g *game) dropGame() bool {
	if len(g.cs) <= 3 {
		return false
	}
	g.t.s[g.cs[len(g.cs)-1]].update(inProcessing, inTable)
	g.cs = g.cs[:len(g.cs)-1]
	return true
}

type gameType int

const (
	kindGame gameType = iota
	seqGame
)

type card int

func newCard(n int, s suit) card {
	return card(n * 4 + int(s))
}

func parseCard(c string) card {
	if len(c) < 3 || len(c) > 4 || (len(c) == 4 && (c[2] < '0' || c[2] > '3')) || c[1] < '1' || c[1] > '9' {
		return 0
	}
	n := c[1] - '0'
	ci := 2
	if c[2] >= '0' && c[2] <= '3' {
		n *= 10
		n += c[2] - '0'
		ci++
	}
	var s suit
	switch c[ci] {
	case 'c':
		s = clubs
	case 'd':
		s = diamonds
	case 'h':
		s = hearts
	case 's':
		s = spades
	default:
		return 0
	}
	return newCard(int(n), s)
}

func (c card) String() string {
	return fmt.Sprintf("%v%v", c.number(), c.suit())
}

func (c card) number() int {
	return int(c) / 4
}

func (c card) suit() suit {
	return suit(int(c) % 4)
}

type suit int

const (
	clubs suit = iota
	diamonds
	hearts
	spades
)

func (s suit) String() string {
	switch s {
	case 0:
		return "c"
	case 1:
		return "d"
	case 2:
		return "h"
	case 3:
		return "s"
	}
	log.Fatal("Invalid suit")
	return ""
}

func insert(cs *[]card, c card) {
	var i int
	for i = 0; i < len(*cs); i++ {
		if (*cs)[i] > c {
			break
		}
	}
	log.Println("insert", *cs, c, i)
	*cs = append((*cs)[:i], append([]card{c}, (*cs)[i:]...)...)
	log.Println("insert", *cs)
}

func remove(cs *[]card, c card) {
	var i int
	for i = 0; i < len(*cs); i++ {
		if (*cs)[i] == c {
			break
		}
	}
	if i == len(*cs) {
		log.Fatal("Remove unexisting")
	}
	*cs = append((*cs)[:i], (*cs)[i+1:]...)
}

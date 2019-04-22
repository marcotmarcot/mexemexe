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
			insert(&m.h, c)
		case 'c':
			// Checks if it's possible to play any card on the hand.
			m.findCard()
		default:
			fmt.Println("Invalid command: ", command)
		}
	}
}

type mexemexe struct {
	t *table
	h []card
}

func newMexemexe() *mexemexe {
	return &mexemexe{newTable(), nil}
}

func (m *mexemexe) findCard() {
	for _, c := range m.h {
		log.Println("findCard", c)
		if m.t.check(c) {
			fmt.Println(c)
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
	defer log.Println(t.s)
	return newProcessing(t).check(0)
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
	log.Println("check", i, len(p.t.cs), c)
	log.Println("p.gs", p.gs)
	if !p.t.s[c].has(inTable) {
		log.Println("InTable")
		return p.check(i + 1)
	}
	if p.buildKind(i) {
		return true
	}
	if p.buildSeq(i) {
		return true
	}
	if c.number() == 1 {
		return p.check(i + 1)
	}
	return false
}

func (p *processing) buildKind(i int) bool {
	c := p.t.cs[i]
	cn := c.number()
	g := newGame(p.t)
	for s := 0; s < 4; s++ {
		nc := newCard(cn, suit(s))
		if !p.t.s[nc].has(inTable) {
			continue
		}
		g.addCard(nc)
	}
	return p.addGame(i, g)
}

func (p *processing) buildSeq(i int) bool {
	c := p.t.cs[i]
	cn := c.number()
	cs := c.suit()
	g := newGame(p.t)
	g.addCard(c)
	for n := cn + 1; n <= 14; n++ {
		en := n
		if en == 14 {
			en = 1
		}
		nc := newCard(en, cs)
		if !p.t.s[nc].has(inTable) {
			break
		}
		g.addCard(nc)
	}
	return p.addGame(i, g)
}

func (p *processing) addGame(i int, g *game) bool {
	if len(g.cs) < 3 {
		g.destroy()
		return false
	}
	p.gs = append(p.gs, g)
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

func newGame(t *table) *game {
	return &game{t, nil}
}

func (g *game) String() string {
	return fmt.Sprintf("%v", g.cs)
}

func (g *game) addCard(c card) {
	g.t.s[c].update(inTable, inProcessing)
	g.cs = append(g.cs, c)
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
	*cs = append((*cs)[:i], append([]card{c}, (*cs)[i:]...)...)
}

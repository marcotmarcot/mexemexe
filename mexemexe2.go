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
				fmt.Println("Invalid command:", command)
				continue
			}
			if err := m.t.newCard(c); err != nil {
				fmt.Println("Invalid command:", command, err)
				continue
			}
		case 'h':
			// Adds the following card to the hand.
			c := parseCard(command)
			if c == 0 {
				fmt.Println("Invalid command:", command)
				continue
			}
			if err := m.h.newCard(c); err != nil {
				fmt.Println("Invalid command:", command, err)
				continue
			}
			// log.Println(m.h.cs)
		case 'c':
			// Checks if it's possible to play any card on the hand.
			cs, err := m.findCard()
			if err != nil {
				fmt.Println("Invalid command:", command, err)
				continue
			}
			if cs == nil {
				fmt.Println("Not found")
				continue
			}
			for _, c := range cs {
				if err := m.h.removeCard(c); err != nil {
					log.Fatal("Invalid command:", command, err)
				}
				if err := m.t.newCard(c); err != nil {
					log.Fatal("Invalid command:", command, err)
				}
			}
			fmt.Println(cs)
		case 'R':
			c := parseCard(command)
			if c == 0 {
				fmt.Println("Invalid command:", command)
				continue
			}
			if err := m.h.removeCard(c); err != nil {
				fmt.Println("Invalid command:", command, err)
				continue
			}
		case 'r':
			c := parseCard(command)
			if c == 0 {
				fmt.Println("Invalid command:", command)
				continue
			}
			if err := m.t.removeCard(c); err != nil {
				fmt.Println("Invalid command:", command, err)
				continue
			}
		default:
			fmt.Println("Invalid command:", command)
		}
		log.Println("Status:", m.t.cs, m.h.cs)
	}
}

type mexemexe struct {
	t *table
	h *table
}

func newMexemexe() *mexemexe {
	return &mexemexe{newTable(), newTable()}
}

func (m *mexemexe) findCard() ([]card, error) {
	for _, c := range m.h.cs {
		if found, err := m.t.check(c); err != nil {
			log.Println("findCard check", c)
			return nil, err
		} else if found {
			return []card{c}, nil
		}
	}
	for i := range m.h.cs {
		if g := m.h.findGame(kindGame, i, true); g != nil {
			return g, nil
		}
		if g := m.h.findGame(seqGame, i, true); g != nil {
			return g, nil
		}
	}
	return nil, nil
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

func (t *table) newCard(c card) error {
	insert(&t.cs, c)
	if err := t.s[c].update(noState, inTable); err != nil {
		log.Println("newCard", c)
		return err
	}
	return nil
}

func (t *table) removeCard(c card) error {
	var i int
	for i = 0; i < len(t.cs); i++ {
		if t.cs[i] == c {
			break
		}
	}
	if i == len(t.cs) {
		return fmt.Errorf("remove %v %v", t.cs, c)
	}
	t.cs = append(t.cs[:i], t.cs[i+1:]...)
	if err := t.s[c].update(inTable, noState); err != nil {
		log.Println("removeCard", c)
		return err
	}
	return nil
}

func (t *table) check(c card) (found bool, rerr error) {
	if err := t.newCard(c); err != nil {
		log.Println("check newCard", c)
		return false, err
	}
	defer func() {
		if err := t.removeCard(c); err != nil {
			log.Println("check removeCard", c)
			rerr = err
		}
	}()
	return newProcessing(t).check(0)
}

func (t *table) findGame(ty gameType, i int, stopAtThree bool) []card {
	var g []card
	c := t.cs[i]
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

func (ss *states) update(from, to state) error {
	for i, s := range ss.ss {
		if s == from {
			ss.ss[i] = to
			return nil
		}
	}
	return fmt.Errorf("No from found %v %v %v", from, to, ss.ss)
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

func (p *processing) check(i int) (bool, error) {
	if i == len(p.t.cs) {
		for _, c := range p.t.cs {
			if p.t.s[c].has(inTable) {
				return false, nil
			}
		}
		log.Println("check", p.gs)
		return true, nil
	}
	c := p.t.cs[i]
	if !p.t.s[c].has(inTable) {
		return p.check(i + 1)
	}
	if found, err := p.buildGame(kindGame, i); err != nil {
		log.Println("check buildGame", kindGame, i)
		return false, err
	} else if found {
		return true, nil
	}
	if found, err := p.buildGame(seqGame, i); err != nil {
		log.Println("check buildGame", seqGame, i)
		return false, err
	} else if found {
		return true, nil
	}
	if c.number() == 1 {
		return p.check(i + 1)
	}
	return false, nil
}

func (p *processing) buildGame(ty gameType, i int) (bool, error) {
	g := p.t.findGame(ty, i, false)
	if g == nil {
		return false, nil
	}
	return p.addGame(i, g)
}

func (p *processing) addGame(i int, cs []card) (found bool, rerr error) {
	g, err := newGame(p.t, cs)
	if err != nil {
		log.Println("addGame", i, cs)
		return false, err
	}
	p.gs = append(p.gs, g)
	defer func() {
		if err := p.destroyGame(); err != nil {
			log.Println("addGame destroyGame", i, cs)
			rerr = err
		}
	}()
	if found, err := p.check(i + 1); err != nil {
		log.Println("addGame check", i, cs)
		return false, err
	} else if found {
		return true, nil
	}
	kept, err := p.dropGame()
	if err != nil {
		log.Println("addGame dropGame", i, cs)
		return false, err
	}
	for kept {
		if found, err := p.check(i + 1); err != nil {
			log.Println("addGame check", i, cs)
			return false, err
		} else if found {
			return true, nil
		}
		kept, err = p.dropGame()
		if err != nil {
			log.Println("addGame kept dropGame", i, cs)
			return false, err
		}
	}
	return false, nil
}

func (p *processing) destroyGame() error {
	g := p.gs[len(p.gs)-1]
	p.gs = p.gs[:len(p.gs)-1]
	return g.destroy()
}

func (p *processing) dropGame() (bool, error) {
	return p.gs[len(p.gs)-1].dropGame()
}

type game struct {
	t *table
	cs []card
}

func newGame(t *table, cs []card) (*game, error) {
	for _, c := range cs {
		if err := t.s[c].update(inTable, inProcessing); err != nil {
			log.Println("newGame", c)
			return nil, err
		}
	}
	return &game{t, cs}, nil
}

func (g *game) String() string {
	return fmt.Sprintf("%v", g.cs)
}

func (g *game) destroy() error {
	for _, c := range g.cs {
		if err := g.t.s[c].update(inProcessing, inTable); err != nil {
			log.Println("destroy", c)
			return err
		}
	}
	return nil
}

func (g *game) dropGame() (bool, error) {
	if len(g.cs) <= 3 {
		return false, nil
	}
	c := g.cs[len(g.cs)-1]
	if err := g.t.s[c].update(inProcessing, inTable); err != nil {
		log.Println("dropGame", c)
		return false, err
	}
	g.cs = g.cs[:len(g.cs)-1]
	return true, nil
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
	if len(c) < 3 || len(c) > 4 || (len(c) == 4 && (c[2] < '0' || c[2] > '3')) || c[1] < '1' || c[1] > '9' || (len(c) == 3 && c[2] >= '0' && c[2] <= '3') {
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

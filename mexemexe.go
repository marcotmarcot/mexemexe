package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
)

func main() {
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
			if c := buildCard(command); c != nil {
				table = append(table, c)
				sort.Sort(cardSlice(table))
				continue
			}
			fmt.Println("Invalid command")
		case 'h':
			// Adds the following card to the hand.
			if c := buildCard(command); c != nil {
				hand = append(hand, c)
				sort.Sort(cardSlice(hand))
				continue
			}
			fmt.Println("Invalid command")
		case 'c':
			// Checks if it's possible to play any card on the hand.
			findCard()
			findGame()
		case 'd':
			fmt.Println(hand, table)
		}
	}
}

func buildCard(c string) *card {
	if c[1] < '1' || c[1] > '9' {
		return nil
	}
	n := c[1] - '0'
	ci := 2
	if c[2] >= '0' && c[2] <= '3' {
		n *= 10
		n += c[2] - '0'
		ci++
	}
	switch c[ci] {
	case 'c', 'd', 'h', 's':
	default:
		return nil
	}
	return &card{n, c[ci]}
}

type card struct {
	n byte
	s byte
}

func (c *card) String() string {
	return strconv.Itoa(int(c.n)) + string([]byte{c.s})
}

var table, hand []*card

type cardSlice []*card

func (c cardSlice) Len() int {
	return len(c)
}

func (c cardSlice) Less(i, j int) bool {
	return c[i].n < c[j].n || c[i].n == c[j].n && c[i].s < c[j].s
}

func (c cardSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func findCard() bool {
	for _, c := range hand {
		t := append([]*card{c}, table...)
		sort.Sort(cardSlice(t))
		if checkTable(t, [][]*card{}, []*card{}, 0) {
			fmt.Println(c)
		}
	}
}

func checkTable(t []*card, games [][]*card, current []*card, li int) bool {
	total := 0
	for _, g := range games {
		total += len(g)
	}
	total += len(current)
	if total == len(t) {
		games = append(games, current)
		fmt.Println(games)
		return true
	}
	for i := li + 1; i < len(t); i++ {
		if len(p) == 1 && !isGamePrefix(p[0], t[i]) {
			continue
		}
		p = append(p, t[i])
		restore := false
		if len(p) > 3 && !isGame(p) {
			p := p[len(p)-1]
			restore := true
			games = append(games, p)
			p = []*card{t[i]}
		}
		if checkTable(t, p, i) {
			return true
		}
		p = p[:len(previous)-1]
		if restore {
			p = games[len(games)-1]
			games = games[:len(games)-1]
		}
	}
	return false
}

func findGame() bool {
	found := false
	var h []*card
	for i, c := range hand {
		if i > 0 && *hand[i - 1] == *hand[i] {
			continue
		}
		h = append(h, c)
	}
	for i1 := 0; i1 < len(h); i1++ {
		c1 := h[i1]
		for i2 := i1 + 1; i2 < len(h); i2++ {
			c2 := h[i2]
			if !isGamePrefix(c1, c2) {
				continue
			}
			for i3 := i2 + 1; i3 < len(h); i3++ {
				c3 := h[i3]
				g := []*card{c1, c2, c3}
				if isGame(g) {
					fmt.Println(g)
					found = true
				}
			}
			if c2.n != 13 {
				continue
			}
			for i3 := 0; i3 < i1 && h[i3].n == 1; i3++ {
				c3 := h[i3]
				g := []*card{c1, c2, c3}
				if isGame(g) {
					fmt.Println(g)
					found = true
				}
			}
		}
	}
	return found
}

func isGamePrefix(c1, c2 *card) bool {
	// 3 of a kind
	return c1.n == c2.n && c1.s != c2.s ||
		// sequence
		follows(c1, c2)
}

func isGame(g []*card) bool {
	// 3 of a kind
	return g[0].n == g[1].n && g[0].n == g[2].n && g[0].s != g[1].s && g[0].s != g[2].s && g[1].s != g[2].s ||
		// sequence
		follows(g[0], g[1]) && followsEnd(g[1], g[2])
}

func follows(c1, c2 *card) bool {
	return c1.s == c2.s && c1.n + 1 == c2.n
}

func followsEnd(c1, c2 *card) bool {
	return c1.s == c2.s && (c1.n + 1 == c2.n) || (c1.n == 13 && c2.n == 1)
}

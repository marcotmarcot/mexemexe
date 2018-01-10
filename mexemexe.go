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
			c := buildCard(command)
			if c == nil {
				fmt.Println("Invalid command")
				continue
			}
			table = append(table, c)
			sort.Sort(cardSlice(table))
		case 'h':
			// Adds the following card to the hand.
			c := buildCard(command)
			if c == nil {
				fmt.Println("Invalid command")
				continue
			}
			hand = append(hand, c)
			sort.Sort(cardSlice(hand))
		case 'p':
			c := buildCard(command)
			if c == nil {
				fmt.Println("Invalid command")
				continue
			}
			var i int
			var h *card
			for i, h = range hand {
				if *h == *c {
					break
				}
			}
			hand = append(hand[:i], hand[i+1:]...)
			table = append(table, c)
			sort.Sort(cardSlice(table))
		case 'r':
			c := buildCard(command)
			if c == nil {
				fmt.Println("Invalid command")
				continue
			}
			var i int
			var t *card
			for i, t = range table {
				if *t == *c {
					break
				}
			}
			table = append(table[:i], table[i+1:]...)
		case 'c':
			// Checks if it's possible to play any card on the hand.
			findCard()
			findGame()
		case 'd':
			fmt.Println(hand, table)
		default:
			fmt.Println("Invalid command")
		}
	}
}

func buildCard(c string) *card {
	if len(c) != 3 || c[1] < '1' || c[1] > '9' {
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

func findCard() {
	for _, c := range hand {
		t := append([]*card{c}, table...)
		sort.Sort(cardSlice(t))
		if checkTable(t, []int{}, 0) {
			fmt.Println(c)
		}
	}
}

func checkTable(t []*card, is []int, current int) bool {
	if len(is) == len(t) {
		var game []*card
		for j := current; j < len(is); j++ {
			game = append(game, t[is[j]])
		}
		if !isGame(game) {
			return false
		}
		var organized []*card
		for _, i := range is {
			organized = append(organized, t[i])
		}
		fmt.Println(organized)
		return true
	}
	for i := range t {
		found := false
		for _, li := range is {
			if i == li {
				found = true
				break
			}
		}
		if found {
			continue
		}
		// fmt.Println(is)
		// fmt.Println(i)
		// fmt.Println(current)
		var game []*card
		for j := current; j < len(is); j++ {
			game = append(game, t[is[j]])
		}
		game = append(game, t[i])
		if len(is) - current == 1 {
			if !isGamePrefix(t[is[len(is)-1]], t[i]) {
				// fmt.Println("GamePrefix", len(is)-1, t[is[len(is)-1]], i, t[i])
				// fmt.Println(t)
				continue
			}
		} else if len(is) - current == 2 {
			if !isGame(game) {
				continue
			}
		} else if len(is) - current > 2 { 
			if !isGame(game) {
				is = append(is, i)
				if checkTable(t, is, len(is)-1) {
					return true
				}
				is = is[:len(is)-1]
				continue
			}
		}
		is = append(is, i)
		if checkTable(t, is, current) {
			return true
		}
		is = is[:len(is)-1]
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
	// 3-4 of a kind
	if len(g) < 3 {
		return false
	}
	kind := true
	seq := true
	for i := 1; i < len(g); i++ {
		if g[i].n != g[0].n || g[i].s == g[0].s {
			kind = false
		}
		if !(follows(g[i-1], g[i]) || i == len(g) - 1 && followsEnd(g[
i-1], g[i])) {
			seq = false
		}
	}
	// fmt.Println(g)
	// fmt.Println("game", kind || seq)
	return kind || seq
}

func follows(c1, c2 *card) bool {
	return c1.s == c2.s && c1.n + 1 == c2.n
}

func followsEnd(c1, c2 *card) bool {
	return c1.s == c2.s && (c1.n + 1 == c2.n) || (c1.n == 13 && c2.n == 1)
}

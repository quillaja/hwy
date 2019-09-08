// formats the 'raw' text to a cleaned up version.
// read io from stdin, output to stdout

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

func parseSubCmd(cmd string) {

	if cmd == "" {
		cmd = "original"
	}

	g := graphFromRaw(os.Stdin)

	keys := byState(g.keys())
	sort.Sort(keys)

	switch cmd {
	case "original":
		// reproduces the formatting of the raw input, but cleaned and sorted
		for _, k := range keys {
			line := append(placeSlice{k}, g[k]...)
			fmt.Println(line)
		}

	case "cities":
		// prints a list of unique places
		for _, k := range keys {
			fmt.Println(k)
		}

	case "readable":
		// prints the graph in a nice way
		for _, k := range keys {
			fmt.Printf("%20s (%d): %s\n", k, len(g[k]), placeSlice(byState(g[k])))

			// columns require too much space, even wrapped
			// fmt.Printf("%19s (%d)", k, len(g[k]))
			// for i, p := range byState(g[k]) {
			// 	if i > 0 && i%3 == 0 {
			// 		fmt.Printf("\n%23s", "") // newline and indent
			// 	}
			// 	fmt.Printf("%20s", p)
			// }
			// fmt.Println()
		}
	}
}

// graphFromRaw reads the raw input from r and produces a graph.
// raw input is a list of places delimited by a semicolon, where the first
// place is the 'vertex' and the following places are connected vertices.
// lines starting with # are comments and are ignored.
func graphFromRaw(r io.Reader) graph {
	g := graph{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		raw := scanner.Text()

		// ignore comments (#)
		if len(raw) == 0 || raw[0] == '#' {
			continue
		}

		rawcities := strings.Split(raw, ";")
		first := parsePlace(rawcities[0])
		if _, ok := g[first]; !ok {
			g[first] = placeSlice{}
		}
		for _, c := range rawcities[1:] {
			g[first] = append(g[first], parsePlace(c))
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	return g
}

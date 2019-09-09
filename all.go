package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

func allSubCmd(cmd string) {
	switch cmd {
	case "file":
		parseData(os.Stdin)

	case "special":
		filled := fromSpecialStdin(os.Stdin)
		fmt.Println(filled.FullString())
	}
}

// demands the 1st line be the number of vertices (cities),
// followed by that number of `city@loc` lines,
// followed by that number of `city;city~weight;...` lines.
// outputs that number of "Full" lines in format:
// `city@loc;city@loc~weight;...`
func fromSpecialStdin(r io.Reader) distGraph {
	s := bufio.NewScanner(r)

	var n int
	s.Scan()
	fmt.Sscan(s.Text(), &n)

	buf := new(bytes.Buffer)
	for i := 0; i < n && s.Scan(); i++ {
		line := s.Text()
		buf.WriteString(line + "\n")
	}
	if s.Err() != nil {
		panic(s.Err())
	}

	places := getLocsFromReader(buf)

	buf.Reset()
	for i := 0; i < n && s.Scan(); i++ {
		line := s.Text()
		buf.WriteString(line + "\n")
	}
	if s.Err() != nil {
		panic(s.Err())
	}

	weights := getDistFromReader(buf)

	num := len(places)
	find := func(other place) place {
		return places[sort.Search(num, func(i int) bool {
			return strings.Compare(places[i].String(), other.String()) >= 0 // bullshit
		})]
	}
	sort.Sort(byCity(places))

	filled := distGraph{} // new distgraph with loc-filled places
	for ki, vi := range weights {
		// fmt.Printf("%3d %s == %s\n", i, ki.FullString(), places[find(ki)].FullString())
		dm := distMap{} // new distmap with loc-filled keys
		for ji, w := range vi {
			dm[find(ji)] = w
		}
		filled[find(ki)] = dm
	}

	return filled
}

// normal representation of verticies and edges.
//     <vertex>;<edge>;<edge>...
// where vertex:
//     <city> <state>@<lat> <lon>
// and edge:
//     <city> <state>@<lat> <lon>~<meters>,<time>
func parseData(r io.Reader) distGraph {
	s := bufio.NewScanner(r)

	for s.Scan() {
		line := s.Text()
		if line == "" {
			continue
		}
		places := strings.Split(line, ";")

		vertex := parseFullPlace(places[0]) // 0 is only place, no weight
		fmt.Println(vertex.FullString())
		for _, str := range places[1:] {
			// split place and weight
			parts := strings.Split(str, "~")
			edge := parseFullPlace(parts[0])
			weight := parseWeight(parts[1])
			fmt.Printf("\t%s\t%s\n", edge.FullString(), weight)
		}
		fmt.Println()

	}

	return distGraph{}
}

func parseFullPlace(pstr string) place {
	parts := strings.Split(pstr, "@")
	vertex := parsePlace(parts[0])
	fmt.Sscan(parts[1], &vertex.Lat, &vertex.Lon)
	return vertex
}

func parseWeight(wstr string) (w weight) {
	var dstr string
	fmt.Sscanf(wstr, "%f,%s", &w.Distance, &dstr)
	w.Time, _ = time.ParseDuration(dstr)
	return
}

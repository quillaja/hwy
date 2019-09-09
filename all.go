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
	case "format2":
		fmt.Print(parseData(os.Stdin).Format2())

	case "special":
		filled := fromSpecialStdin(os.Stdin)
		fmt.Print(filled.FullString())

	case "finaltypes":
		g := ParseGraph(os.Stdin)
		for k, v := range g {
			fmt.Printf("\n%s, %s (%g, %g)\n", k.City, k.State, k.Latitude, k.Longitude)
			for kk, vv := range v {
				fmt.Printf("\t%-16s%5.1fmi\n", kk.City, vv.Distance*MetersToMiles)
			}
		}

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

	dg := distGraph{}
	for s.Scan() {
		line := s.Text()
		if line == "" {
			continue
		}
		places := strings.Split(line, ";")

		vertex := parseFullPlace(places[0]) // 0 is only place, no weight
		// fmt.Println(vertex.FullString())
		dm := distMap{}
		for _, str := range places[1:] {
			// split place and weight
			parts := strings.Split(str, "~")
			edge := parseFullPlace(parts[0])
			weight := parseWeight(parts[1])
			dm[edge] = weight
			// fmt.Printf("\t%s\t%s\n", edge.FullString(), weight)
		}
		dg[vertex] = dm
		// fmt.Println()

	}

	return dg
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

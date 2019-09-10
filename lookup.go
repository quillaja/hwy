package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func lookupSubCmd(cmd string) {
	switch cmd {
	case "name":
		g := ParseGraph(os.Stdin)
		name := argN(3, "none")

		fmt.Println("search1")
		places := g.Places()
		sort.Sort(ByCity(places))
		index := sort.Search(len(places), func(i int) bool {
			return strings.Compare(name, places[i].City) <= 0
		})
		if index < len(places) {
			found := places[index]
			fmt.Println(found.String())
		} else {
			fmt.Println("didn't find: ", name)
		}

		fmt.Println("search2")
		parts := strings.Split(name, minorSep) // requires comma separated
		fmt.Println(g.FindPlace(parts[0], parts[1]))

	case "loc":
		g := ParseGraph(os.Stdin)
		lat, _ := strconv.ParseFloat(argN(3, "0"), 64)
		lon, _ := strconv.ParseFloat(argN(4, "0"), 64)
		dist, _ := strconv.ParseFloat(argN(5, "0"), 64)
		fmt.Println(lat, lon, dist)
		fmt.Println(g.FindWithin(lat, lon, dist))
	}
}

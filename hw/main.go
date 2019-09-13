package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/quillaja/hwy"
)

var apikey string

func main() {
	cmd := argN(1, "")

	b, err := ioutil.ReadFile("KEY")
	if err != nil {
		panic(err)
	}
	apikey = string(b)

	switch cmd {

	case "find":
		switch argN(2, "") {
		case "name":
			g := hwy.ParseGraph(os.Stdin)
			name := argN(3, "none")

			parts := strings.Split(name, ",") // requires comma separated
			fmt.Println(g.FindPlace(parts[0], parts[1]))

		case "loc":
			g := hwy.ParseGraph(os.Stdin)
			lat, _ := strconv.ParseFloat(argN(3, "0"), 64)
			lon, _ := strconv.ParseFloat(argN(4, "0"), 64)
			dist, _ := strconv.ParseFloat(argN(5, "0"), 64)
			fmt.Println(lat, lon, dist)
			fmt.Println(g.FindWithin(lat, lon, dist))

		case "path":
			g := hwy.ParseGraph(os.Stdin)
			origIn := strings.Split(argN(3, ""), ",") // assume "CITY NAME,STATE"
			destIn := strings.Split(argN(4, ""), ",")
			orig, _ := g.FindPlace(origIn[0], origIn[1])
			dest, _ := g.FindPlace(destIn[0], destIn[1])
			fmt.Printf("shortest path between %s and %s:\n", orig.Name(), dest.Name())
			path := g.ShortestPath(orig, dest)
			fmt.Println(path)
		}
	case "pipeline":
		hwy.ConvertRaw(os.Stdin, os.Stdout, apikey)

	case "check":
		fmt.Println("undirected =", hwy.RawIsUndirected(os.Stdin))
	}
}

func argN(n int, def string) string {
	if len(os.Args) > n {
		return os.Args[n]
	}
	return def
}

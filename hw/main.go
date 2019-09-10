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
		}
	case "pipeline":
		hwy.ConvertRaw(os.Stdin, os.Stdout, apikey)
	}
}

func argN(n int, def string) string {
	if len(os.Args) > n {
		return os.Args[n]
	}
	return def
}

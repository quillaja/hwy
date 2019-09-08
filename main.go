package main

import (
	"io/ioutil"
	"os"
)

var apikey string

func main() {
	cmd := argN(1, "parse")

	b, err := ioutil.ReadFile("KEY")
	if err != nil {
		panic(err)
	}
	apikey = string(b)

	switch cmd {
	case "parse":
		parseSubCmd(argN(2, ""))

	case "loc":
		latlonSubCmd(argN(2, ""))

	case "dist":
		distSubCmd(argN(2, ""))
	}
}

func argN(n int, def string) string {
	if len(os.Args) > n {
		return os.Args[n]
	}
	return def
}

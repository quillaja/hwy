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

	case "all":
		allSubCmd(argN(2, ""))

	case "lookup":
		lookupSubCmd(argN(2, ""))

	case "pipeline":
		ConvertRaw(os.Stdin, os.Stdout, apikey)
	}
}

func argN(n int, def string) string {
	if len(os.Args) > n {
		return os.Args[n]
	}
	return def
}

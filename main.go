package main

import "os"

func main() {
	cmd := argN(1, "parse")

	switch cmd {
	case "parse":
		parseSubCmd(argN(2, ""))
	}
}

func argN(n int, def string) string {
	if len(os.Args) > n {
		return os.Args[n]
	}
	return def
}

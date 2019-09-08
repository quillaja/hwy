package main

import "os"

func main() {
	cmd := argN(1, "parse")

	switch cmd {
	case "parse":
		parseSubCmd(argN(2, ""))

	case "latlon":
		apikey, ok := os.LookupEnv("GMAPS_KEY")
		if !ok {
			return
		}
		latlonSubCmd(apikey)

	case "dist":

		apikey, ok := os.LookupEnv("GMAPS_KEY")
		if !ok {
			return
		}
		distSubCmd(apikey)
	}
}

func argN(n int, def string) string {
	if len(os.Args) > n {
		return os.Args[n]
	}
	return def
}

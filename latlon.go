package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"googlemaps.github.io/maps"
)

func latlonSubCmd(cmd string) {

	var places placeSlice

	switch cmd {
	case "request":
		g := graphFromRaw(os.Stdin)
		places = g.keys()
		sort.Sort(byState(places))

		getLocsFromGoogle(&places)

		// print
		for _, p := range places {
			fmt.Println(p.FullString())
		}

	case "file":
		places := placeSlice{}
		getLocsFromReader(os.Stdin, &places)

		// print
		for _, p := range places {
			fmt.Printf("%s | ", p.FullString())
		}
	}
}

func getLocsFromGoogle(places *placeSlice) {
	client, err := maps.NewClient(maps.WithAPIKey(apikey))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	for i, p := range *places {
		req := &maps.GeocodingRequest{Address: p.String()}
		results, err := client.Geocode(ctx, req)
		if err != nil {
			fmt.Println(p, err)
		}

		if len(results) < 1 {
			fmt.Println(p, "no results")
			continue
		}

		(*places)[i].lat = results[0].Geometry.Location.Lat
		(*places)[i].lon = results[0].Geometry.Location.Lng

		if len(results) > 1 {
			fmt.Println(p, "more than 1 result?", results)
		}
	}
}

func getLocsFromReader(r io.Reader, pspointer *placeSlice) {
	places := placeSlice{}
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		parts := strings.Split(line, "@")
		p := parsePlace(parts[0])
		fmt.Sscanf(parts[1], "%f %f", &p.lat, &p.lon)
		places = append(places, p)
	}
	*pspointer = places // overwrite whatever was passed in
}

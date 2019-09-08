package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"googlemaps.github.io/maps"
)

func latlonSubCmd(cmd string) {
	apikey := cmd

	client, err := maps.NewClient(maps.WithAPIKey(apikey))
	if err != nil {
		panic(err)
	}

	g := graphFromRaw(os.Stdin)
	places := g.keys()
	sort.Sort(byState(places))

	ctx := context.Background()
	for i, p := range places {
		req := &maps.GeocodingRequest{Address: p.String()}
		results, err := client.Geocode(ctx, req)
		if err != nil {
			fmt.Println(p, err)
		}

		if len(results) < 1 {
			fmt.Println(p, "no results")
			continue
		}

		places[i].lat = results[0].Geometry.Location.Lat
		places[i].lon = results[0].Geometry.Location.Lng

		if len(results) > 1 {
			fmt.Println(p, "more than 1 result?", results)
		}
	}

	// print
	for _, p := range places {
		fmt.Printf("%s@%f %f\n", p, p.lat, p.lon)
	}
}

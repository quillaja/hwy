package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"googlemaps.github.io/maps"
)

// These types and functions process a "raw" file (just place names)
// into a complete graph with locations and weights.

// graphFromRaw reads the raw input from r and produces a graph.
// raw input is a list of places delimited by a semicolon, where the first
// place is the 'vertex' and the following places are connected vertices.
// lines starting with # are comments and are ignored.

type rawPlace struct {
	city, state string
}

type rawGraph map[rawPlace][]rawPlace

// parseRawPlace parses raw format `city name state` into a place. Raw format
// always has last 2 chars as the state abbreviation.
func parseRawPlace(raw string) (p rawPlace) {
	split := len(raw) - 2
	p.city = strings.Title(strings.TrimSpace(raw[:split]))
	p.state = strings.ToUpper(strings.TrimSpace(raw[split:]))
	return
}

func rawKeys(rg rawGraph) []rawPlace {
	rp := make([]rawPlace, len(rg))
	for k := range rg {
		rp = append(rp, k)
	}
	return rp
}

func parseRawGraph(r io.Reader) rawGraph {
	g := rawGraph{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		raw := scanner.Text()

		// ignore comments (#)
		if len(raw) == 0 || raw[0] == '#' {
			continue
		}

		rawcities := strings.Split(raw, ";")
		first := parseRawPlace(rawcities[0])
		if _, ok := g[first]; !ok {
			g[first] = []rawPlace{}
		}
		for _, c := range rawcities[1:] {
			g[first] = append(g[first], parseRawPlace(c))
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	return g
}

func requestLocsFromGoogle(raw []rawPlace, apikey string) []Place {
	places := make([]Place, len(raw))

	client, err := maps.NewClient(maps.WithAPIKey(apikey))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	for i, p := range raw {
		req := &maps.GeocodingRequest{Address: fmt.Sprintf("%s, %s", p.city, p.state)}
		results, err := client.Geocode(ctx, req)
		if err != nil {
			fmt.Println(p, err)
		}

		if len(results) < 1 {
			fmt.Println(p, "no results")
			continue
		}

		places[i].Latitude = results[0].Geometry.Location.Lat
		places[i].Longitude = results[0].Geometry.Location.Lng

		if len(results) > 1 {
			fmt.Println(p, "more than 1 result?", results)
		}
	}

	return places
}

func convertRawGraphToGraph(raw rawGraph, places []Place) Graph {
	g := make(Graph, len(raw))

	// use 'find' to easily get Place from rawPlace
	sort.Sort(ByCity(places))
	find := func(rp rawPlace) Place {
		i := sort.Search(len(places), func(i int) bool {
			return strings.Compare(rp.city+","+rp.state, places[i].Name()) <= 0
		})
		return places[i]
	}

	for k, v := range raw {
		em := make(EdgeMap, len(v)) // create an empty edge map
		for _, p := range v {
			// convert each rawPlace to Place and insert into EdgeMap
			em[find(p)] = Weight{} // zero-val Weight
		}
		g[find(k)] = em // add the origin and edges
	}

	return g
}

// modifies graph in-place, returns same one
func requestDistFromGoogle(g Graph, apikey string) Graph {
	client, err := maps.NewClient(maps.WithAPIKey(apikey))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	for orig, dests := range g {

		// prepare request
		req := &maps.DistanceMatrixRequest{Origins: []string{orig.Name()}}

		destIndex := make(map[Place]int, len(dests)) // need this so we can get 'elements' in correct order later
		i := 0
		for d := range dests {
			destIndex[d] = i
			i++
			req.Destinations = append(req.Destinations, d.Name())
		}

		// make request
		resp, err := client.DistanceMatrix(ctx, req)
		if err != nil {
			fmt.Println(orig, err)
			continue
		}
		if len(resp.Rows) == 0 {
			fmt.Println(orig, "no rows")
			continue
		}
		if len(resp.Rows) > 1 {
			fmt.Println(orig, "more than one row")
		}

		// fill distMap with values
		for d, dest := range dests {
			elem := resp.Rows[0].Elements[destIndex[d]]
			dest.Distance = float64(elem.Distance.Meters)
			dest.TravelTime = elem.Duration
			dests[d] = dest
		}
		// put the filled distMap back into the original graph
		g[orig] = dests
	}

	return g
}

func FullyProcessRaw(r io.Reader, apikey string) Graph {
	rg := parseRawGraph(r)
	places := requestLocsFromGoogle(rawKeys(rg), apikey) // makes calls to Google Maps Geocoding API
	g := convertRawGraphToGraph(rg, places)
	g = requestDistFromGoogle(g, apikey) // makes calls to Google Maps Distance Matrix API
	return g
}

package hwy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

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

func (p rawPlace) String() string {
	return p.city + minorSep + p.state
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
	rp := make([]rawPlace, 0, len(rg))
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
		neighbors := []rawPlace{}
		for _, c := range rawcities[1:] {
			neighbors = append(neighbors, parseRawPlace(c))
		}
		if _, ok := g[first]; !ok {
			g[first] = neighbors
		} else {
			fmt.Fprintln(os.Stderr, first, "already was in raw graph.")
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	return g
}

func rawIsUndirected(g rawGraph) (undirected bool) {
	// check that every node is in the adjacency list of all it's neighbors
	undirected = true
	for a, aadj := range g {
		for _, b := range aadj {
			in := false
			for _, badj := range g[b] {
				if a == badj {
					in = true
					break
				}
			}
			if !in {
				fmt.Fprintf(os.Stderr, "%s -> %s but NOT %s -> %s\n", a, b, b, a)
				undirected = false
			}
		}
	}
	return
}

func requestLocsFromGoogle(raw []rawPlace, apikey string) []Place {
	places := make([]Place, 0, len(raw))

	client, err := maps.NewClient(maps.WithAPIKey(apikey))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	for _, p := range raw {
		req := &maps.GeocodingRequest{Address: fmt.Sprintf("%s, %s", p.city, p.state)}
		results, err := client.Geocode(ctx, req)
		if err != nil {
			fmt.Fprintln(os.Stderr, p, err)
		}

		if len(results) < 1 {
			fmt.Fprintln(os.Stderr, p, "no results")
			continue
		}

		places = append(places, Place{
			City:      p.city,
			State:     p.state,
			Latitude:  results[0].Geometry.Location.Lat,
			Longitude: results[0].Geometry.Location.Lng,
		})

		if len(results) > 1 {
			fmt.Fprintln(os.Stderr, p, "more than 1 result?", results)
		}
	}

	return places
}

func convertRawGraphToGraph(raw rawGraph, places []Place) Graph {
	g := make(Graph, len(raw))

	// use 'find' to easily get Place from rawPlace
	sort.Sort(ByCity(places))
	find := func(rp rawPlace) Place {
		// linear search
		var i int
		for ; i < len(places); i++ {
			if places[i].Name() == rp.city+","+rp.state {
				break
			}
		}
		// TODO: reinstate this (binary search)
		// i := sort.Search(len(places), func(i int) bool {
		// 	return strings.Compare(rp.city+","+rp.state, places[i].Name()) <= 0
		// })
		if i >= len(places) {
			fmt.Fprintln(os.Stderr, i, rp, places)
		}
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

		// need to keep a list of destinations in the same order as was used
		// to prepare the request, so that the response data can be retrieved
		// in the correct order.
		destlist := make([]Place, 0, len(dests))
		for d := range dests {
			destlist = append(destlist, d)
			req.Destinations = append(req.Destinations, d.Name())
		}

		// make request
		resp, err := client.DistanceMatrix(ctx, req)
		if err != nil {
			fmt.Fprintln(os.Stderr, orig, err)
			continue
		}
		if len(resp.Rows) == 0 {
			fmt.Fprintln(os.Stderr, orig, "no rows")
			continue
		}
		if len(resp.Rows) > 1 {
			fmt.Fprintln(os.Stderr, orig, "more than one row")
		}

		// add the Weight data to Graph
		for i, d := range destlist {
			elem := resp.Rows[0].Elements[i]
			g[orig][d] = Weight{Distance: float64(elem.Distance.Meters), TravelTime: elem.Duration}
		}
	}

	return g
}

// FullyProcessRaw reads and parses the "raw hand-entered data" format for
// highway connections, then uses Google APIs to get place locations (lat, lon)
// and travel times and distances (by car) between connected places. Checks
// graph for undirectedness before performing requests and returns nil (as well
// as printing errors to Stderr) if the graph has directed edges.
//
// NOTE: the Google Maps API requires a key and may also incure useage charges.
func FullyProcessRaw(r io.Reader, apikey string) Graph {
	fmt.Fprintln(os.Stderr, "parsing raw data from input Reader.")
	rg := parseRawGraph(r)

	if !rawIsUndirected(rg) {
		return nil
	}

	fmt.Fprintf(os.Stderr, "requesting locations from Google. %d Geocoding API calls.\n", len(rg))
	places := requestLocsFromGoogle(rawKeys(rg), apikey) // makes calls to Google Maps Geocoding API

	fmt.Fprintf(os.Stderr, "got back %d places.\n", len(places))
	fmt.Fprintln(os.Stderr, "adding locations to graph.")
	g := convertRawGraphToGraph(rg, places)

	fmt.Fprintf(os.Stderr, "requesting distances from Google. %d Distance Matrix API calls.\n", len(g))
	g = requestDistFromGoogle(g, apikey) // makes calls to Google Maps Distance Matrix API

	return g
}

// ConvertRaw does the same steps as FullyProcessRaw(), but writes the resulting
// graph to w instead of returning it.
func ConvertRaw(r io.Reader, w io.Writer, apikey string) {
	start := time.Now()
	fmt.Fprintln(os.Stderr, "starting.")

	g := FullyProcessRaw(r, apikey)

	fmt.Fprintln(os.Stderr, "writing fully processed graph to output Writer.")
	w.Write([]byte(g.String()))

	fmt.Fprintf(os.Stderr, "done. took %s.\n", time.Since(start))
}

// RawIsUndirected parses raw graph data from r and checks that
// each vertex (place) is in the adjacency list of its neighbors. Returns
// true if so, false if not. Prints undirected edges to Stderr.
func RawIsUndirected(r io.Reader) bool {
	rg := parseRawGraph(r)
	return rawIsUndirected(rg)
}

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"googlemaps.github.io/maps"
)

func distSubCmd(cmd string) {

	switch cmd {
	case "request":
		g := graphFromRaw(os.Stdin)
		dg := makeDistGraph(g)

		getDistFromGoogle(&dg)

		fmt.Println(dg)

	case "file":
		fmt.Println(getDistFromReader(os.Stdin))

	case "json":
		return
		g := graphFromRaw(os.Stdin)
		dg := makeDistGraph(g)

		getDistFromGoogle(&dg)

		b, err := json.MarshalIndent(dg, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	}
}

func getDistFromGoogle(dm *distGraph) {
	client, err := maps.NewClient(maps.WithAPIKey(apikey))
	if err != nil {
		panic(err)
	}

	// count := 0
	ctx := context.Background()
	for orig, dests := range *dm {
		// prepare request
		req := &maps.DistanceMatrixRequest{
			Origins: []string{orig.String()}}
		destIndex := map[place]int{} // need this so we can get 'elements' in correct order later
		i := 0
		for d := range dests {
			destIndex[d] = i
			i++
			req.Destinations = append(req.Destinations, d.String())
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
			dest.Time = elem.Duration
			dests[d] = dest
		}
		// put the filled distMap back into the original graph
		(*dm)[orig] = dests

		// testing
		// if count > 5 {
		// 	break
		// }
		// count++
	}
}

func getDistFromReader(r io.Reader) distGraph {
	dg := distGraph{}
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		cities := strings.Split(line, ";")
		key := parsePlace(cities[0])
		dm := distMap{}
		for _, raw := range cities[1:] {
			parts := strings.Split(raw, "~")
			d := parsePlace(parts[0])
			w := weight{}
			t := ""
			fmt.Sscanf(parts[1], "%f,%s", &w.Distance, &t)
			w.Time, _ = time.ParseDuration(t)
			dm[d] = w
		}
		dg[key] = dm
	}

	return dg
}

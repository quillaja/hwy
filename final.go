package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	majorSep = ";"
	minorSep = ","
)

// Conversion multipiers.
const (
	// Meters in 1 mile.
	MilesToMeters = 1609.344

	// Miles in 1 meter.
	MetersToMiles = 1 / MilesToMeters
)

// Place is an 'origin' or 'destination' city, given by it's name, state,
// and geographic coordinates.
type Place struct {
	City      string
	State     string
	Latitude  float64
	Longitude float64
}

func (p Place) String() string {
	// City,State,Lat,Lon
	return fmt.Sprintf("%[1]s%[5]s%[2]s%[5]s%[3]f%[5]s%[4]f", p.City, p.State, p.Latitude, p.Longitude, minorSep)
}

// Weight is a type to hold edge data: the travel distance in meters and time
// in time.Duration by car.
type Weight struct {
	Distance   float64 // meters
	TravelTime time.Duration
}

func (w Weight) String() string {
	// Distance,TravelTime
	return fmt.Sprintf("%[1]g%[3]s%[2]s", w.Distance, w.TravelTime, minorSep)
}

// EdgeMap contains 'destination' Places as keys and the
// travel distance and time as values
type EdgeMap map[Place]Weight

func (em EdgeMap) String() string {
	strs := make([]string, 0, len(em))
	for p, w := range em {
		// place,weight
		strs = append(strs, fmt.Sprintf("%[1]s%[3]s%[2]s", p, w, minorSep))
	}
	return strings.Join(strs, majorSep)
}

// Graph is the primary data structure for holding the highway data. The map's
// key is the 'origin' Place, and the EdgeMap contains 'destination' Places
// as keys and the travel distance and time as values.
type Graph map[Place]EdgeMap

func (g Graph) String() string {
	b := strings.Builder{}
	for vertex, edges := range g {
		// vertex;edge
		b.WriteString(fmt.Sprintf("%[1]s%[3]s%[2]s\n", vertex, edges, majorSep))
	}
	return b.String()
}

// PrettyPrint writes a nicely formatted tabular list of places and distances
// between them to w. The resulting output is many lines long.
func (g Graph) PrettyPrint(w io.Writer) {
	newline := ""
	for k, v := range g {
		fmt.Fprintf(w, "%s%s, %s (%g, %g)\n", newline, k.City, k.State, k.Latitude, k.Longitude)
		newline = "\n" // redundant but this avoids if
		for kk, vv := range v {
			fmt.Fprintf(w, "\t%-16s%3s%7.1fmi%10s\n", kk.City, kk.State, vv.Distance*MetersToMiles, vv.TravelTime)
		}
	}
}

// Places provides a slice of unique Places in the graph.
// Use `ByState` or `ByCity` with pkg `sort` to sort the slice.
func (g Graph) Places() []Place {
	places := make([]Place, 0, len(g))
	for k := range g {
		places = append(places, k)
	}
	return places
}

// Edge gets the Weight (edge data) for the connection between origin and
// destination if it exists. If the two places are not connected, data is
// the zero value and ok is false.
func (g Graph) Edge(origin, destination Place) (data Weight, ok bool) {
	if dm, ok := g[origin]; ok {
		if data, ok := dm[destination]; ok {
			return data, true
		}
	}
	return
}

//
//
//
// The following few types are used with Graph.Most()
//
//
//

// MinMax is a type of function that returns a min or max (or...?)
// of two floats.
type MinMax func(float64, float64) float64

// Predefined MinMaxes for Graph.Most().
var (
	// The max of 2 floats.
	Max MinMax = math.Max

	// The min of 2 floats.
	Min MinMax = math.Min
)

// Accessor ins a function that 'converts' a Weight to a float.
type Accessor func(Weight) float64

// Predefined Accessors for Graph.Most().
var (
	// Gets Weight.Distance in meters.
	Dist Accessor = func(w Weight) float64 { return w.Distance }

	// Gets Weight.TravelTime in minutes.
	Time Accessor = func(w Weight) float64 { return w.TravelTime.Minutes() }
)

// Most will find the "mostest" edge of vertex 'origin' given the predicate and
// Accessor 'by'. Generally, it'll return the farthest or closest city connected
// to 'origin' based on the value of Weight specified with 'by'.
//
// For example `g.Most(myHomeTown, Max, Dist)` returns the farthest by distance
// connected city to `myHomeTown`.
//
// ok is false if origin is not found in the Graph.
func (g Graph) Most(origin Place, predicate MinMax, by Accessor) (most Place, ok bool) {
	dm, ok := g[origin]
	if !ok {
		return most, false
	}

	var cur *Weight
	for k, v := range dm {
		// start by setting the first value to the current best
		if cur == nil {
			most = k
			*cur = v
			continue
		}

		// get the mostest
		bestval := predicate(by(*cur), by(v))
		if bestval != by(*cur) {
			// the other was chosen
			most = k
			*cur = v
		}
	}

	return most, true
}

// ParseGraph parses input from r, successively turning each line into a new
// entry in the graph. Lines beginning with "#" are ignored ascomments, and
// blank lines are skipped. Line format is:
// `<place:city,state,lat,lon>;<place>,<weight:distance,time>;<place>,<weight>;...`
func ParseGraph(r io.Reader) Graph {
	s := bufio.NewScanner(r)

	g := Graph{}
	for s.Scan() {
		line := s.Text()
		// skip blank and comment
		if len(line) == 0 || strings.TrimSpace(string(line[0])) == "#" {
			continue
		}

		parts := strings.Split(line, majorSep)
		vertex := ParsePlace(parts[0])
		edges := EdgeMap{}
		for _, part := range parts[1:] {
			dest := ParsePlace(part) // this will work on strings with "extra" fields
			dparts := strings.Split(part, minorSep)
			w := Weight{} // TODO: refactor
			w.Distance, _ = strconv.ParseFloat(dparts[4], 64)
			w.TravelTime, _ = time.ParseDuration(dparts[5])
			edges[dest] = w
		}
		g[vertex] = edges
	}

	return g
}

// ParsePlace parses a Place from a string in the format:
// `city,state,latitude,longitude`
func ParsePlace(str string) (p Place) {
	parts := strings.Split(str, minorSep)
	p.City = parts[0]
	p.State = parts[1]
	p.Latitude, _ = strconv.ParseFloat(parts[2], 64)
	p.Longitude, _ = strconv.ParseFloat(parts[3], 64)
	return
}

//
//
//
// sorting []Place
//
//
//

// ByState allows sorting []Place by state then city name.
type ByState []Place

// Len is the number of elements in the collection.
func (p ByState) Len() int {
	return len(p)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (p ByState) Less(i int, j int) bool {
	cmp := strings.Compare(p[i].State, p[j].State)
	if cmp == 0 {
		// same state, compare city
		return strings.Compare(p[i].City, p[j].City) == -1
	}
	return cmp == -1
}

// Swap swaps the elements with indexes i and j.
func (p ByState) Swap(i int, j int) {
	p[i], p[j] = p[j], p[i]
}

// ByCity allows sorting []Place by city name then state.
type ByCity []Place

// Len is the number of elements in the collection.
func (p ByCity) Len() int {
	return len(p)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (p ByCity) Less(i int, j int) bool {
	cmp := strings.Compare(p[i].City, p[j].City)
	if cmp == 0 {
		// same city, compare state
		return strings.Compare(p[i].State, p[j].State) == -1
	}
	return cmp == -1
}

// Swap swaps the elements with indexes i and j.
func (p ByCity) Swap(i int, j int) {
	p[i], p[j] = p[j], p[i]
}

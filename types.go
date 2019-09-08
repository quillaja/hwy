package main

import (
	"fmt"
	"strings"
	"time"
)

// parsePlace parses raw format `city name state` into a place. Raw format
// always has last 2 chars as the state abbreviation.
func parsePlace(raw string) (p place) {
	split := len(raw) - 2
	p.City = strings.Title(strings.TrimSpace(raw[:split]))
	p.State = strings.ToUpper(strings.TrimSpace(raw[split:]))
	return
}

// place holds the city data
type place struct {
	City, State string
	Lat, Lon    float64
}

// Equal compares places only by city and state (ignores lat lon).
func (p place) Equal(other place) bool {
	return p.City == other.City && p.State == other.State
}

func (p place) String() string {
	return fmt.Sprintf("%s %s", p.City, p.State)
}

func (p place) FullString() string {
	return fmt.Sprintf("%s %s@%f %f", p.City, p.State, p.Lat, p.Lon)
}

func (p place) displayWidth() int {
	return len(p.City) + 3
}

// sort a place slice by state.
type byState placeSlice

// Len is the number of elements in the collection.
func (p byState) Len() int {
	return len(p)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (p byState) Less(i int, j int) bool {
	cmp := strings.Compare(p[i].State, p[j].State)
	if cmp == 0 {
		// same state, compare city
		return strings.Compare(p[i].City, p[j].City) == -1
	}
	return cmp == -1
}

// Swap swaps the elements with indexes i and j.
func (p byState) Swap(i int, j int) {
	p[i], p[j] = p[j], p[i]
}

// sort a place slice by city name.
type byCity placeSlice

// Len is the number of elements in the collection.
func (p byCity) Len() int {
	return len(p)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (p byCity) Less(i int, j int) bool {
	cmp := strings.Compare(p[i].City, p[j].City)
	if cmp == 0 {
		// same city, compare state
		return strings.Compare(p[i].State, p[j].State) == -1
	}
	return cmp == -1
}

// Swap swaps the elements with indexes i and j.
func (p byCity) Swap(i int, j int) {
	p[i], p[j] = p[j], p[i]
}

// a slice of places.
type placeSlice []place

func (p placeSlice) String() string {
	strs := make([]string, len(p))
	for i := range p {
		strs[i] = p[i].String()
	}
	return strings.Join(strs, "; ")
}

// graph associates a place with places connected to it.
type graph map[place]placeSlice

// gets the keys of the graph (and therefore a set of unique places).
func (g graph) keys() placeSlice {
	keys := placeSlice{}
	for k := range g {
		keys = append(keys, k)
	}
	return keys
}

type weight struct {
	Distance float64 // meters
	Time     time.Duration
}

func (w weight) String() string {
	return fmt.Sprintf("%f,%s", w.Distance, w.Time)
}

type distMap map[place]weight

func (dm distMap) String() string {
	strs := []string{}
	for k, v := range dm {
		strs = append(strs, fmt.Sprintf("%s~%s", k, v))
	}
	return strings.Join(strs, ";")
}

func (dm distMap) FullString() string {
	strs := []string{}
	for k, v := range dm {
		strs = append(strs, fmt.Sprintf("%s~%s", k.FullString(), v))
	}
	return strings.Join(strs, ";")
}

type distGraph map[place]distMap

func makeDistGraph(g graph) (dg distGraph) {
	dg = distGraph{}

	for k, v := range g {
		m := distMap{}
		for _, p := range v {
			m[p] = weight{}
		}
		dg[k] = m
	}
	return
}

func (dg distGraph) String() string {
	builder := strings.Builder{}
	for k, v := range dg {
		builder.WriteString(fmt.Sprintf("%s;%s\n", k, v))
	}
	return builder.String()
}

func (dg distGraph) FullString() string {
	builder := strings.Builder{}
	for k, v := range dg {
		builder.WriteString(fmt.Sprintf("%s;%s\n", k.FullString(), v.FullString()))
	}
	return builder.String()
}

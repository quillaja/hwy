package main

import (
	"fmt"
	"strings"
)

// parsePlace parses raw format `city name state` into a place. Raw format
// always has last 2 chars as the state abbreviation.
func parsePlace(raw string) (p place) {
	split := len(raw) - 2
	p.city = strings.Title(strings.TrimSpace(raw[:split]))
	p.state = strings.ToUpper(strings.TrimSpace(raw[split:]))
	return
}

// place holds the city data
type place struct {
	city, state string
	// lat, lon, elev float64
}

func (p place) String() string {
	return fmt.Sprintf("%s %s", p.city, p.state)
}

func (p place) displayWidth() int {
	return len(p.city) + 3
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
	cmp := strings.Compare(p[i].state, p[j].state)
	if cmp == 0 {
		// same state, compare city
		return strings.Compare(p[i].city, p[j].city) == -1
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
	cmp := strings.Compare(p[i].city, p[j].city)
	if cmp == 0 {
		// same city, compare state
		return strings.Compare(p[i].state, p[j].state) == -1
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

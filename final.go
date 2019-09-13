package hwy

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

// String returns a single string with all the Place's attributes, no spaces,
// separated using `minorSep` (usually comma).
func (p Place) String() string {
	// City,State,Lat,Lon
	return fmt.Sprintf("%[1]s%[5]s%[2]s%[5]s%[3]f%[5]s%[4]f", p.City, p.State, p.Latitude, p.Longitude, minorSep)
}

// Name returns a single string with the Place's city and state, no spaces,
// separated using `minorSep` (usually comma).
func (p Place) Name() string {
	return fmt.Sprintf("%s%s%s", p.City, minorSep, p.State)
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

// String returns a single string with the full data of the graph.
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

//
//
//
// Search
//
//
//

// FindPlace does a linear search of Places and returns the first match.
// If found, match is the place and found is true. Otherwise match is the
// zero value and found is false.
func (g Graph) FindPlace(city, state string) (match Place, found bool) {
	city = strings.ToLower(city)
	state = strings.ToLower(state)
	for p := range g {
		if strings.ToLower(p.City) == city &&
			strings.ToLower(p.State) == state {
			return p, true
		}
	}

	return Place{}, false
}

// FindWithin performs a linear search of Places and returns the closest match
// that is within `radius` meters of the given latitude and longitude. The function
// uses the "spherical law of cosines" to calculate the distance. `found` is
// false if no Place was found.
func (g Graph) FindWithin(lat, lon, radius float64) (match Place, dist float64, found bool) {
	best := radius + 1
	for p := range g {
		d := sphericalLawOfCos(lat, lon, p.Latitude, p.Longitude)
		if d <= radius && d < best {
			match = p
			dist = d
			found = true
			best = d
		}
	}

	if !found {
		return Place{}, 0, false
	}
	return
}

// used for sphericalLawOfCos()
const (
	earthradius = 6371e3 // 6371 km = 6,371,000 m
	degtorad    = math.Pi / 180.0
)

// sphericalLawOfCos uses said law to calculate the distance in meters (because
// `earthradius` is in meters) between (lat1,lon1) and (lat2,lon2).
//
// d = acos( sin φ1 ⋅ sin φ2 + cos φ1 ⋅ cos φ2 ⋅ cos Δλ ) ⋅ R
func sphericalLawOfCos(lat1, lon1, lat2, lon2 float64) float64 {
	lat1 *= degtorad
	lat2 *= degtorad
	return earthradius * math.Acos(
		math.Sin(lat1)*math.Sin(lat2)+
			math.Cos(lat1)*math.Cos(lat2)*
				math.Cos((lon2-lon1)*degtorad))
}

//
//
// Dijkstra's algorithm
//
//

type PathMap map[Place]pdata

type pdata struct {
	visited bool
	Dist    float64
	Hops    int
	parent  Place
}

func (pm PathMap) Path(dest Place) (path []Place, sum float64) {
	// prepare path if applicable
	hops := pm[dest].Hops
	if hops > 0 { // hops==0 -> no path found
		path = make([]Place, hops+1, hops+1) // +1 to include origin in path
		// build reverse path
		// for n := dest; n != none; n = pm[n].parent {
		// 	path = append(path, n)
		// }
		// // swap all into correct order
		// for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		// 	path[i], path[j] = path[j], path[i]
		// }
		n := dest
		for i := len(path) - 1; i >= 0; i-- {
			path[i] = n
			n = pm[n].parent
		}
		sum = pm[dest].Dist
	}
	return
}

// ShortestPath finds the shortest paths between orig and all other vertices
// using Dijkstra's algorithm.
//
// https://en.wikipedia.org/wiki/Dijkstra%27s_algorithm
func (g Graph) ShortestPath(orig Place, by Accessor) PathMap {

	inf := math.Inf(1)
	none := Place{} // zero val
	var d pdata     // temp var for data

	// 1. mark all nodes unvisitied. create a set of all unvisited nodes
	// call the unvisited set
	// 2. assign to every node a tentative distance value: zero for initial node
	// and infinity ("unvisited") for all others. Set initial node as current.
	nodes := make(PathMap, len(g))
	for k := range g {
		nodes[k] = pdata{Dist: inf}
	}

	current := orig
	d = nodes[current]
	d.Dist = 0
	nodes[current] = d

	found := false // aka done

	for !found {
		// fmt.Println("current", current, nodes[current])
		if current == none {
			return nil
		}

		// 3. for the current node, consider all its unvisited neighbors and
		// calculate their tentative distances through the current node. Compare
		// the newly calculated tentative distance to the currently assigned value
		// and assign the smaller value.
		for n, w := range g[current] {
			if !nodes[n].visited { // n in unvisited set
				tentative := nodes[current].Dist + by(w)
				d = nodes[n]
				if d.Dist > tentative {
					d.Dist = tentative
					d.parent = current
					d.Hops = nodes[d.parent].Hops + 1
					nodes[n] = d
				}
			}
		}

		// 4. when we are done considering all the unvisited neighbors of the
		// current node, mark the current node as visited and remove it from the
		// unvisited set. A visited node will never be checked again.
		d = nodes[current]
		d.visited = true
		nodes[current] = d

		// 5. A) if all nodes are marked visited (unvisited set is empty)
		// OR B) if the smallest tentative distance among nodes in the unvisited set
		// is infinity (no path possible)
		// The algorithm is finished.
		// TODO: termination case B
		unvisitedcount := 0
		for _, d := range nodes {
			if !d.visited {
				unvisitedcount++
			}
		}

		found = unvisitedcount == 0
		if found {
			continue
		}

		// 6. Otherwise, select the unvisited node that is marked with the smallest
		// tentative value, set it as the "current" and go back to step 3.
		minDist := inf // pos infinity
		minPlace := Place{}
		for node, d := range nodes {
			if !d.visited && d.Dist < minDist {
				minDist = d.Dist
				minPlace = node
			}
		}
		current = minPlace
		found = minDist == inf // termination case 5B above
	}

	return nodes
}

//
//
// parsing Graph and Place
//
//

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

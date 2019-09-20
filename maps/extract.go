// +build ignore

package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//
//
// types for states.xml or countries.xml
//
//

type StateList struct {
	States []State `xml:"state"`
}

type State struct {
	ID       string    `xml:"id,attr"`
	Name     string    `xml:"statename,attr"`
	Polygons []Polygon `xml:"polygon"`
}

func (state *State) Parse() {
	for i := range state.Polygons {
		state.Polygons[i].ParsePoints()
	}
}

func (state *State) Shapes() (shape [][]Point) {
	shape = make([][]Point, 0, len(state.Polygons))
	for i := range state.Polygons {
		shape = append(shape, state.Polygons[i].Points)
	}
	return
}

type Polygon struct {
	Raw    string `xml:"points,attr"`
	Points []Point
}

func (poly *Polygon) ParsePoints() []Point {
	r := strings.NewReader(poly.Raw)
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanWords)

	poly.Points = []Point{}
	var p Point
	for s.Scan() {
		fmt.Sscanf(s.Text(), "%f,%f", &p[1], &p[0])
		poly.Points = append(poly.Points, p)
	}
	kill(s.Err())

	return poly.Points
}

type Point [2]float64

//
//
// type for cities.xml
//
//

type rawCityList struct {
	Cities []rawCity `xml:"city"`
}

type rawCity struct {
	ID         string  `xml:"id,attr"`
	Population int     `xml:"population,attr"`
	Capital    string  `xml:"capital,attr"`
	Latitude   float64 `xml:"lat,attr"`
	Longitude  float64 `xml:"lon,attr"`
}

type City struct {
	Name       string
	Country    string
	Population int
	Capital    bool
	Latitude   float64
	Longitude  float64
}

func (c City) StringList() []string {
	s := make([]string, 6)
	s[0] = c.Name
	s[1] = c.Country
	s[2] = strconv.Itoa(c.Population)
	s[3] = strconv.FormatBool(c.Capital)
	s[4] = strconv.FormatFloat(c.Latitude, 'f', 3, 64)
	s[5] = strconv.FormatFloat(c.Longitude, 'f', 3, 64)
	return s
}

func CityFromStringList(fields []string) City {
	c := City{
		Name:    fields[0],
		Country: fields[1]}

	c.Population, _ = strconv.Atoi(fields[2])
	c.Capital, _ = strconv.ParseBool(fields[3])
	c.Latitude, _ = strconv.ParseFloat(fields[4], 64)
	c.Longitude, _ = strconv.ParseFloat(fields[5], 64)

	return c
}

//
//
// program
//
//

func main() {

	if len(os.Args) < 3 {
		fmt.Println("USEAGE: extract TYPE FILENAME\n TYPE=(state, city)")
		os.Exit(1)
	}

	filetype := os.Args[1]
	filename := os.Args[2]
	path := strings.TrimSuffix(filename, filepath.Ext(filename))

	data, err := ioutil.ReadFile(filename)
	kill(err)

	switch filetype {
	case "state":

		var doc StateList
		err = xml.Unmarshal(data, &doc)
		kill(err)

		for _, s := range doc.States {
			s.Parse()
			for name, style := range formatters {
				// create path for the data <filename>/<format>
				fmtpath := filepath.Join(path, name)
				os.MkdirAll(fmtpath, 0755)

				// create file and output data
				fname := strings.ToLower(strings.ReplaceAll(s.ID, " ", "_"))
				file, err := os.Create(filepath.Join(fmtpath, fname+"."+name))
				kill(err)

				style(file, &s)

				file.Close()
			}
		}

	case "city":

		var doc rawCityList
		err = xml.Unmarshal(data, &doc)
		kill(err)

		// remake the xml 'raw city' structs to the nicer kind
		cities := make([]City, 0, len(doc.Cities))
		for _, c := range doc.Cities {
			parts := strings.Split(c.ID, ",")
			cities = append(cities, City{
				Name:       strings.TrimSpace(parts[0]),
				Country:    strings.TrimSpace(parts[1]),
				Capital:    c.Capital == "Y",
				Population: c.Population,
				Latitude:   c.Latitude,
				Longitude:  c.Longitude,
			})
		}

		// create directory for the data files
		os.MkdirAll(path, 0755)
		for name, style := range cityformatters {
			// create data file for this format (1 file per format)
			file, err := os.Create(filepath.Join(path, "cities."+name))
			kill(err)

			// write cities in format
			style(file, cities)

			file.Close()
		}

	}
}

func kill(err error) {
	if err != nil {
		panic(err)
	}
}

type format func(io.Writer, *State)

func plaintext(w io.Writer, state *State) {
	// first line is ID STATE_NAME NUM_POLYGONS
	// followed by NUM_POLYGONS groups of points making the polygon.
	// each polygon is then NUM_POINTS on a line
	// then NUM_POINTS number of lines, each consisting of 2 floats (lat and lon)
	// TODO: replace empty string ID or Name with space
	fmt.Fprintf(w, "%s %s %d\n",
		strings.ReplaceAll(state.ID, " ", "_"),
		strings.ReplaceAll(state.Name, " ", "_"),
		len(state.Polygons))

	for _, p := range state.Polygons {
		fmt.Fprintln(w, len(p.Points))
		for i := range p.Points {
			fmt.Fprintf(w, "%.6f %.6f\n", p.Points[i][0], p.Points[i][1])
		}
	}
}

func jsontext(w io.Writer, state *State) {
	// cleaned up data struct
	data := struct {
		ID       string
		Name     string
		Polygons [][]Point
	}{
		ID:       state.ID,
		Name:     state.Name,
		Polygons: state.Shapes(),
	}

	enc := json.NewEncoder(w)
	// enc.SetIndent("", " ")
	kill(enc.Encode(data))
}

var formatters = map[string]format{
	"txt":  plaintext,
	"json": jsontext,
}

//
// city formatters
//

type cityformat func(io.Writer, []City)

func cityplaintext(w io.Writer, cities []City) {
	// CSV format
	enc := csv.NewWriter(w)
	for i := range cities {
		kill(enc.Write(cities[i].StringList()))
	}
}

func cityjsontext(w io.Writer, cities []City) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	kill(enc.Encode(cities))
}

var cityformatters = map[string]cityformat{
	"csv":  cityplaintext,
	"json": cityjsontext,
}

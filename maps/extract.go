package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Doc struct {
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

func main() {

	data, err := ioutil.ReadFile("states.xml")
	kill(err)

	var doc Doc
	err = xml.Unmarshal(data, &doc)
	kill(err)

	for _, s := range doc.States {
		s.Parse()
		for name, style := range formatters {
			file, err := os.Create(filepath.Join(name, s.ID))
			kill(err)
			style(file, &s)
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
	fmt.Fprintf(w, "%s %s %d\n",
		state.ID, strings.ReplaceAll(state.Name, " ", "_"), len(state.Polygons))
	for _, p := range state.Polygons {
		fmt.Fprintln(w, len(p.Points))
		for i := range p.Points {
			fmt.Fprintf(w, "%.6f %.6f\n", p.Points[i][0], p.Points[i][1])
		}
	}
}

func jsontext(w io.Writer, state *State) {
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
	"plaintext": plaintext,
	"json":      jsontext,
}

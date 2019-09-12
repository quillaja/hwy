package maps

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

//
//
// types for state or country data
//
//

type State struct {
	ID       string
	Name     string
	Polygons [][]Point
}

type Point [2]float64

func StateFromText(r io.Reader) State {
	// format
	// ID NAME NUM_POLYGONS
	// NUM_POINTS
	// LAT LON
	// ...
	s := State{}

	// read number of polygons and make slice
	var numPoly int
	fmt.Fscanln(r, &s.ID, &s.Name, &numPoly)
	s.Polygons = make([][]Point, numPoly, numPoly)

	for i := range s.Polygons {
		// read number of points and make point slice
		var numPoints int
		fmt.Fscanln(r, &numPoints)
		pts := make([]Point, numPoints, numPoints)

		for j := range pts {
			fmt.Fscanln(r, &pts[j][0], &pts[j][1])
		}
		s.Polygons[i] = pts
	}

	// put the previously replaced spaces back
	s.ID = strings.ReplaceAll(s.ID, "_", " ")
	s.Name = strings.ReplaceAll(s.Name, "_", " ")

	return s
}

func StateFromJSON(r io.Reader) State {
	s := State{}
	dec := json.NewDecoder(r)
	dec.Decode(&s)
	return s
}

//
//
// types for cities data
//
//

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

func CitiesFromJSON(r io.Reader) []City {
	var c []City
	dec := json.NewDecoder(r)
	dec.Decode(&c)
	return c
}

func CitiesFromCSV(r io.Reader) []City {
	cities := []City{}

	dec := csv.NewReader(r)
	dec.ReuseRecord = true

	var rec []string
	var err error
	for {
		rec, err = dec.Read()
		if err != nil {
			break
		}
		cities = append(cities, CityFromStringList(rec))
	}

	return cities
}

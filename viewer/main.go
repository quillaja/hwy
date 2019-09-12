package main

import (
	"os"

	"github.com/quillaja/hwy"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/quillaja/goutil/pxu"
	"github.com/quillaja/hwy/maps"
	"golang.org/x/image/colornames"
)

func kill(err error) {
	if err != nil {
		panic(err)
	}
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "hwy viewer",
		Bounds: pixel.R(0, 0, 1200, 800),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// get point data
	file, err := os.Open("maps/countries/json/united_states.json")
	kill(err)
	usa := maps.StateFromJSON(file)
	file.Close()

	const mapscale = 20
	// draw points into pixel object
	shape := imdraw.New(nil)
	shape.Color = colornames.Black
	for i := range usa.Polygons {
		for j := range usa.Polygons[i] {
			// x = longitude, y = latitude (reversed from "normal")
			shape.Push(pixel.Vec{usa.Polygons[i][j][1], usa.Polygons[i][j][0]}.Scaled(mapscale))
		}
		shape.Polygon(0.2)
	}

	// get graph data
	file, err = os.Open("data/data")
	kill(err)
	graph := hwy.ParseGraph(file)
	file.Close()

	// draw graph verticies
	vertices := imdraw.New(nil)
	vertices.Color = colornames.Blue
	for _, place := range graph.Places() {
		vertices.Push(pixel.Vec{place.Longitude, place.Latitude}.Scaled(mapscale))
		vertices.Circle(0.8, 0)
	}
	// draw graph edges
	edges := imdraw.New(nil)
	edges.Color = colornames.Red
	for orig, dests := range graph {
		for dest := range dests {
			edges.Push(pixel.Vec{orig.Longitude, orig.Latitude}.Scaled(mapscale))
			edges.Push(pixel.Vec{dest.Longitude, dest.Latitude}.Scaled(mapscale))
			edges.Line(0.2)
		}
	}

	// make grid
	grid := imdraw.New(nil)
	grid.Color = colornames.Gray
	const xmax = 200
	const ymax = 90
	for x := -xmax; x <= xmax; x += 10 {
		grid.Push(pixel.Vec{float64(x), ymax}.Scaled(mapscale))
		grid.Push(pixel.Vec{float64(x), -ymax}.Scaled(mapscale))
		grid.Line(1)
	}
	for y := -ymax; y <= ymax; y += 10 {
		grid.Push(pixel.Vec{xmax, float64(y)}.Scaled(mapscale))
		grid.Push(pixel.Vec{-xmax, float64(y)}.Scaled(mapscale))
		grid.Line(0.2)
	}

	// make camera control
	cam := pxu.NewMouseCamera(win.Bounds().Center())
	cam.XExtents.High = 180 * mapscale
	cam.XExtents.Low = -180 * mapscale
	cam.YExtents.High = 90 * mapscale
	cam.YExtents.Low = -90 * mapscale
	// cam.Position = pixel.V(-70*mapscale, 45*mapscale)

	for !win.Closed() {

		if win.JustPressed(pixelgl.KeyHome) {
			cam.Reset()
		}

		cam.Update(win)
		win.SetMatrix(cam.GetMatrix())

		win.Clear(colornames.White)
		grid.Draw(win)
		shape.Draw(win)
		edges.Draw(win)
		vertices.Draw(win)

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

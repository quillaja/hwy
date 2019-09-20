package main

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"time"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/faiface/pixel/text"

	"github.com/quillaja/hwy"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/quillaja/goutil/pxu"
	"github.com/quillaja/hwy/maps"
	"golang.org/x/image/colornames"
)

// display parameters
const (
	pointSearchDist = 10e3 // 10km radius
	mapThickness    = 0.1
	edgeThickness   = 0.05
	gridThickness   = 0.1

	mapscale   = 10.0
	labelscale = 0.1
)

func kill(err error) {
	if err != nil {
		panic(err)
	}
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:     "hwy viewer",
		Bounds:    pixel.R(0, 0, 1000, 1000),
		VSync:     true,
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	mmatrix := pixel.IM.Scaled(pixel.ZV, mapscale)

	// get point data
	file, err := os.Open("maps/countries/json/united_states.json")
	kill(err)
	usa := maps.StateFromJSON(file)
	file.Close()

	// draw points into pixel object
	shape := imdraw.New(nil)
	shape.Color = colornames.Black
	shape.SetMatrix(mmatrix)
	for i := range usa.Polygons {
		for j := 0; j < len(usa.Polygons[i]); j += 10 { // draw every Nth point
			// x = longitude, y = latitude (reversed from "normal")
			shape.Push(pixel.Vec{usa.Polygons[i][j][1], usa.Polygons[i][j][0]}) //.Scaled(mapscale))
		}
		shape.Polygon(mapThickness)
	}

	// get graph data
	file, err = os.Open("data/data")
	kill(err)
	graph := hwy.ParseGraph(file)
	file.Close()

	// draw graph verticies
	vertices := imdraw.New(nil)
	vertices.Color = colornames.Blue
	vertices.SetMatrix(mmatrix)

	font, _ := truetype.Parse(goregular.TTF)
	face := truetype.NewFace(font, &truetype.Options{Size: 32})
	atlas := text.NewAtlas(face, text.ASCII)
	labels := text.New(pixel.ZV, atlas)
	labels.Color = colornames.Black

	for _, place := range graph.Places() {
		p := pixel.Vec{place.Longitude, place.Latitude} //.Scaled(mapscale)
		vertices.Push(p)
		vertices.Circle(pointSearchDist*mTd(place.Latitude), 0)

		labels.Dot = p.Scaled(mapscale / labelscale).Add(pixel.V(1, 1))
		labels.WriteString(place.Name())
	}
	// draw graph edges
	edges := imdraw.New(nil)
	edges.Color = colornames.Red
	edges.SetMatrix(mmatrix)
	for orig, dests := range graph {
		for dest := range dests {
			edges.Push(pixel.Vec{orig.Longitude, orig.Latitude}) //.Scaled(mapscale))
			edges.Push(pixel.Vec{dest.Longitude, dest.Latitude}) //.Scaled(mapscale))
			edges.Line(edgeThickness)
		}
	}

	// make grid
	grid := imdraw.New(nil)
	grid.Color = colornames.Gray
	grid.SetMatrix(mmatrix)
	const xmax = 180
	const ymax = 90
	for x := -xmax; x <= xmax; x += 10 {
		grid.Push(pixel.Vec{float64(x), ymax})  //.Scaled(mapscale))
		grid.Push(pixel.Vec{float64(x), -ymax}) //.Scaled(mapscale))
		grid.Line(gridThickness)
	}
	for y := -ymax; y <= ymax; y += 10 {
		grid.Push(pixel.Vec{xmax, float64(y)})  //.Scaled(mapscale))
		grid.Push(pixel.Vec{-xmax, float64(y)}) //.Scaled(mapscale))
		grid.Line(gridThickness)
	}

	// make camera control
	cam := pxu.NewMouseCamera(win.Bounds().Center())
	cam.XExtents.High = 180 * mapscale
	cam.XExtents.Low = -180 * mapscale
	cam.YExtents.High = 90 * mapscale
	cam.YExtents.Low = -90 * mapscale
	cam.ZExtents.High *= mapscale
	cam.ZExtents.Low *= 1 / mapscale
	// cam.Position = pixel.V(-90*mapscale, 38*mapscale)

	overlay := NewPathOverlay(colornames.Lime, mmatrix)
	pathtype := &hwy.Dist

	for !win.Closed() && !win.JustPressed(pixelgl.KeyEscape) {

		if win.JustPressed(pixelgl.KeyHome) {
			cam.Reset()
		}
		if win.JustPressed(pixelgl.KeySpace) {
			if pathtype == &hwy.Dist {
				pathtype = &hwy.Time
				fmt.Println("Shortest paths using Travel Time.")
			} else {
				pathtype = &hwy.Dist
				fmt.Println("Shortest paths using Distance.")
			}
		}
		if win.JustPressed(pixelgl.MouseButtonMiddle) {
			p := cam.Unproject(win.MousePosition()).Scaled(1 / mapscale)
			fmt.Printf("<clk @ (%f, %f)>\n", p.Y, p.X)
		}
		if win.JustPressed(pixelgl.MouseButtonRight) {
			p := cam.Unproject(win.MousePosition()).Scaled(1 / mapscale)
			place, dist, found := graph.FindWithin(p.Y, p.X, pointSearchDist)

			if found {
				pp, pm := overlay.Push(place, graph.ShortestPath(place, *pathtype))
				if pp == nil {
					fmt.Println(place, dist)
				} else {
					var cost string
					switch pathtype {
					case &hwy.Dist:
						cost = fmt.Sprintf("%.2fmi", pm[place].Dist*hwy.MetersToMiles)
					case &hwy.Time:
						cost = fmt.Sprintf("%s", time.Duration(pm[place].Dist)*time.Minute)
					}
					fmt.Printf("\t%s to %s: %s\tpassing through %d cities.\n",
						pp.Name(), place.Name(), cost, pm[place].Hops-1)
				}
			} else {
				overlay.Clear()
			}
		}

		cam.Update(win)
		win.SetMatrix(cam.GetMatrix())

		win.Clear(colornames.White)

		grid.Draw(win)
		shape.Draw(win)
		edges.Draw(win)
		vertices.Draw(win)
		labels.Draw(win, pixel.IM.Scaled(labels.Orig, labelscale))
		overlay.Draw(win)

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

const degtorad = math.Pi / 180.0

// Returns multiplier to convert Degrees to Meters at the given latitude.
func mTd(latitude float64) (degreesPerMeter float64) {
	const mPDegEquator = 111319.9
	return 1 / (mPDegEquator * math.Cos(latitude*degtorad))
}

// PathOverlay is a type to simplify drawing paths.
type PathOverlay struct {
	im  *imdraw.IMDraw
	pts []hwy.Place
	pms []hwy.PathMap
}

// NewPathOverlay creates an Overlay
func NewPathOverlay(c color.Color, matrix pixel.Matrix) *PathOverlay {
	ol := &PathOverlay{
		im:  imdraw.New(nil),
		pts: []hwy.Place{},
		pms: []hwy.PathMap{}}
	ol.im.SetColorMask(c)
	ol.im.SetMatrix(matrix)
	return ol
}

// draws a single dot
func (ol *PathOverlay) point(p hwy.Place) {
	ol.im.Push(pixel.Vec{p.Longitude, p.Latitude})
	ol.im.Circle(pointSearchDist*mTd(p.Latitude), 0)
}

// draws a path
func (ol *PathOverlay) line(path []hwy.Place) {
	// doesn't draw very first place because it assumes it was already drawn
	for i := 0; i < len(path)-1; i++ {
		ol.im.Push(pixel.Vec{path[i].Longitude, path[i].Latitude})
		ol.im.Push(pixel.Vec{path[i+1].Longitude, path[i+1].Latitude})
		ol.im.Line(edgeThickness)
		ol.point(path[i+1])
	}
}

// clears imdraw of graphical things.
func (ol *PathOverlay) clearGraphics() {
	ol.im.Clear()
	ol.im.Reset()
}

// Push enqueues places. Returns a pointer to the previously Pushed place and
// its PathMap (nil if no previous).
func (ol *PathOverlay) Push(pt hwy.Place, pm hwy.PathMap) (prevPlace *hwy.Place, prevPathMap hwy.PathMap) {
	ol.pts = append(ol.pts, pt)
	ol.pms = append(ol.pms, pm)
	n := len(ol.pts) // len(ol.pms) is same

	if n == 1 {
		// if this is the first Place, simply draw the dot
		ol.point(pt)
	} else {
		// otherwise, must draw the paths between each point.
		prev := ol.pms[n-2]
		path, _ := prev.Path(pt)
		ol.line(path)
		prevPlace, prevPathMap = &ol.pts[n-2], prev
	}
	return
}

// Clear removes all the data and drawing primitives from the overlay.
func (ol *PathOverlay) Clear() {
	// then slice down to len 0. no memory leaks since no references are being held
	ol.pts = ol.pts[:0]
	ol.pms = ol.pms[:0]

	ol.clearGraphics()
}

func (ol *PathOverlay) Draw(win pixel.Target) {
	ol.im.Draw(win)
}

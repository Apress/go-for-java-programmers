package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

// Default game history.
var CoreGame = &Game{
	make(map[string]*GameRun),
	10,
	0,
	1}

// Represents a game.
type Game struct {
	Runs           map[string]*GameRun
	MaxCycles      int
	SkipCycles     int // not currently used
	GoroutineCount int
}

// Run a set of cycles from the grid defined by an image.
func (g *Game) Run(name, url string) (err error) {
	gr, err := NewGameRun(name, url, g)
	if err != nil {
		return
	}
	g.Runs[gr.Name] = gr
	err = gr.Run()
	return
}

// Clear a game.
func (g *Game) Clear() {
	for k, _ := range g.Runs {
		delete(g.Runs, k)
	}
}

// Represents a single run of a game.
type GameRun struct {
	Parent         *Game
	Name           string
	ImageURL       string
	StartedAt      time.Time
	EndedAt        time.Time
	Width, Height  int
	InitialGrid    *Grid
	CurrentGrid    *Grid
	FinalGrid      *Grid
	Cycles         []*GameCycle
	DelayIn10ms    int
	PlayIndex      int
	GoroutineCount int
}

// B & W color indexes
const (
	offIndex = 0
	onIndex  = 1
)

// B & W color palette
var paletteBW = []color.Color{color.White, color.Black}

// Generate a PNG result (single frame).
func (gr *GameRun) MakePNG(writer io.Writer, index int) (err error) {
	var grid *Grid
	switch index {
	case 0:
		grid = gr.InitialGrid
	default:
		index--
		if index < 0 || index >= len(gr.Cycles) {
			err = BadIndexError
			return
		}
		grid = gr.Cycles[index].AfterGrid
	}
	mag := magFactorFlag
	rect := image.Rect(0, 0, mag*gr.Width+1, mag*gr.Height+1)
	img := image.NewPaletted(rect, paletteBW)
	gr.FillImage(grid, img)
	b, err := gr.encodePNGImage(img)
	if err != nil {
		return
	}
	count, err := writer.Write(b.Bytes())
	log.Printf("Returned PNG, size= %d\n", count)
	if saveImageFlag {
		saveFile := fmt.Sprintf("/temp/Image_%s_%d.png", gr.Name, index)
		xerr := ioutil.WriteFile(saveFile, b.Bytes(), os.ModePerm)
		fmt.Printf("Save %s: %v\n", saveFile, xerr)
	}
	return
}

// Make a PNG image.
func (gr *GameRun) encodePNGImage(img *image.Paletted) (b bytes.Buffer, err error) {
	var e png.Encoder
	e.CompressionLevel = png.NoCompression
	err = e.Encode(&b, img)
	return
}

// Generate a GIF result (>= 1 frame).
func (gr *GameRun) MakeGIFs(count int) (agif *gif.GIF, err error) {
	mag := magFactorFlag
	cycles := len(gr.Cycles)
	xcount := cycles + 1
	if xcount > count {
		xcount = count
	}
	added := 0
	agif = &gif.GIF{LoopCount: 5}

	rect := image.Rect(0, 0, mag*gr.Width+1, mag*gr.Height+1)
	img := image.NewPaletted(rect, paletteBW)
	if added < xcount {
		gr.AddGrid(gr.InitialGrid, img, agif)
		added++
	}
	for i := 0; i < cycles; i++ {
		if added < xcount {
			img = image.NewPaletted(rect, paletteBW)
			gc := gr.Cycles[i]
			grid := gc.AfterGrid
			gr.AddGrid(grid, img, agif)
			added++
		}
	}
	return
}

// Fill in and record a cycle image in an animated GIF.
func (gr *GameRun) AddGrid(grid *Grid, img *image.Paletted, agif *gif.GIF) {
	gr.FillImage(grid, img)
	agif.Image = append(agif.Image, img)
	agif.Delay = append(agif.Delay, gr.DelayIn10ms)
}

// Fill in an image from a grid.
func (gr *GameRun) FillImage(grid *Grid, img *image.Paletted) {
	mag := magFactorFlag
	for row := 0; row < grid.Height; row++ {
		for col := 0; col < grid.Width; col++ {
			index := offIndex
			if grid.getCell(col, row) != 0 {
				index = onIndex
			}
			// apply magnification
			for i := 0; i < mag; i++ {
				for j := 0; j < mag; j++ {
					img.SetColorIndex(mag*row+i, mag*col+j, uint8(index))
				}
			}
		}
	}
}

const midValue = 256 / 2 // middle color value

// Error values.
var (
	NotPNGError   = errors.New("not a png")
	NotRGBAError  = errors.New("not RGBA color")
	BadIndexError = errors.New("bad index")
)

// Start a new game run.
func NewGameRun(name, url string, parent *Game) (gr *GameRun, err error) {
	gr = &GameRun{}
	gr.Parent = parent
	gr.Name = name
	gr.GoroutineCount = CoreGame.GoroutineCount
	gr.ImageURL = url
	gr.DelayIn10ms = 5 * 100
	var img image.Image
	var kind string
	img, kind, err = LoadImage(url)
	if err != nil {
		return
	}
	fmt.Printf("Image kind:  %v\n", kind)
	if kind != "png" {
		return nil, NotPNGError
	}
	bounds := img.Bounds()
	minX, minY, maxX, maxY := bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y
	size := bounds.Size()
	//xsize := size.X * size.Y
	gr.InitialGrid = NewEmptyGrid(size.X, size.Y)
	gr.Width = gr.InitialGrid.Width
	gr.Height = gr.InitialGrid.Height

	err = gr.InitGridFromImage(minX, maxX, minY, maxY, img)
	if err != nil {
		return
	}
	gr.CurrentGrid = gr.InitialGrid.DeepCloneGrid()
	return
}

// Fill in a grid from an image.
// Map color images to B&W.  Only RGBA images allowed.
func (gr *GameRun) InitGridFromImage(minX, maxX, minY, maxY int,
	img image.Image) (err error) {
	setCount, totalCount := 0, 0
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			//			r, g, b, a := img.At(x, y).RGBA()
			rgba := img.At(x, y)
			var r, g, b uint8
			switch v := rgba.(type) {
			case color.NRGBA:
				r, g, b, _ = v.R, v.G, v.B, v.A
			case color.RGBA:
				r, g, b, _ = v.R, v.G, v.B, v.A
			default:
				err = NotRGBAError
				return
			}
			cv := byte(0) // assume cell dead
			if int(r)+int(g)+int(b) < midValue*3 {
				cv = byte(1) // make cell alive
				setCount++
			}
			gr.InitialGrid.setCell(x, y, cv)
			totalCount++
		}
	}
	return
}

// Play a game.
// Run requested cycle count.
func (gr *GameRun) Run() (err error) {
	gr.StartedAt = time.Now()
	for count := 0; count < gr.Parent.MaxCycles; count++ {
		err = gr.NextCycle()
		if err != nil {
			return
		}
	}
	gr.EndedAt = time.Now()
	fmt.Printf("GameRun total time: %dms, goroutine count: %d\n",
		(gr.EndedAt.Sub(gr.StartedAt)+NanosPerMs)/NanosPerMs, gr.GoroutineCount)
	gr.FinalGrid = gr.CurrentGrid.DeepCloneGrid()
	return
}

// Represents a single cycle of a game.
type GameCycle struct {
	Parent     *GameRun
	Cycle      int
	StartedAt  time.Time
	EndedAt    time.Time
	BeforeGrid *Grid
	AfterGrid  *Grid
}

func NewGameCycle(parent *GameRun) (gc *GameCycle) {
	gc = &GameCycle{}
	gc.Parent = parent
	return
}

// Advance and play next game cycle.
// Updating of cycle grid rows can be done in parallel;
// which can reduce execution time.
func (gr *GameRun) NextCycle() (err error) {
	gc := NewGameCycle(gr)
	gc.BeforeGrid = gr.CurrentGrid.DeepCloneGrid()
	p := gc.Parent
	goroutineCount := p.Parent.GoroutineCount
	if goroutineCount <= 0 {
		goroutineCount = 1
	}
	gc.AfterGrid = NewEmptyGrid(gc.BeforeGrid.Width, gc.BeforeGrid.Height)
	gc.StartedAt = time.Now()
	// process rows across  allowed goroutines
	rowCount := (gr.Height + goroutineCount/2) / goroutineCount
	var wg sync.WaitGroup
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go processRows(&wg, gc, rowCount, i*rowCount, gc.BeforeGrid, gc.AfterGrid)
	}
	wg.Wait() // let all finish
	gc.EndedAt = time.Now()
	gr.CurrentGrid = gc.AfterGrid.DeepCloneGrid()
	gr.Cycles = append(gr.Cycles, gc)
	gc.Cycle = len(gr.Cycles)
	return
}

// Represents a 2-dimensional game grid (abstract, not as an image).
type Grid struct {
	Data          []byte
	Width, Height int
}

func NewEmptyGrid(w, h int) (g *Grid) {
	g = &Grid{}
	g.Data = make([]byte, w*h)
	g.Width = w
	g.Height = h
	return
}

func (g *Grid) DeepCloneGrid() (c *Grid) {
	c = &Grid{}
	lg := len(g.Data)
	c.Data = make([]byte, lg, lg)
	for i, b := range g.Data {
		c.Data[i] = b
	}
	c.Width = g.Width
	c.Height = g.Height
	return
}

func (g *Grid) getCell(x, y int) (b byte) {
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return
	}
	return g.Data[x+y*g.Width]
}
func (g *Grid) setCell(x, y int, b byte) {
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return
	}
	g.Data[x+y*g.Width] = b
}

// Play game as subset of grid rows (so can be done in parallel).
func processRows(wg *sync.WaitGroup, gc *GameCycle, rowCount int,
	startRow int, inGrid, outGrid *Grid) {
	defer wg.Done()
	gr := gc.Parent
	for index := 0; index < rowCount; index++ {
		rowIndex := index + startRow
		for colIndex := 0; colIndex < gr.Width; colIndex++ {
			// count any neighbors
			neighbors := 0
			if inGrid.getCell(colIndex-1, rowIndex-1) != 0 {
				neighbors++
			}
			if inGrid.getCell(colIndex, rowIndex-1) != 0 {
				neighbors++
			}
			if inGrid.getCell(colIndex+1, rowIndex-1) != 0 {
				neighbors++
			}
			if inGrid.getCell(colIndex-1, rowIndex) != 0 {
				neighbors++
			}
			if inGrid.getCell(colIndex+1, rowIndex) != 0 {
				neighbors++
			}
			if inGrid.getCell(colIndex-1, rowIndex+1) != 0 {
				neighbors++
			}
			if inGrid.getCell(colIndex, rowIndex+1) != 0 {
				neighbors++
			}
			if inGrid.getCell(colIndex+1, rowIndex+1) != 0 {
				neighbors++
			}

			// determine next generation cell state based on neighbor count
			pv := inGrid.getCell(colIndex, rowIndex)
			nv := uint8(0) // assume dead
			switch neighbors {
			case 2:
				nv = pv // unchanged
			case 3:
				if pv == 0 {
					nv = 1 // make alive
				}
			}
			outGrid.setCell(colIndex, rowIndex, nv)
		}
	}
}

package main

import (
	"errors"
	"fmt"
	"math"
)

var ErrBadRange = errors.New("bad range")

type PlotFunc func(in float64) (out float64)

// Print (to STDOUT) the plots of one or more functions.
func PlotPrinter(xsteps, ysteps int, xmin, xmax, ymin, ymax float64,
	fs ...PlotFunc) (err error) {
	xdiff, ydiff := xmax-xmin, ymax-ymin
	if xdiff <= 0 || ydiff <= 0 {
		err = ErrBadRange
		return
	}
	xstep, ystep := xdiff/float64(xsteps), ydiff/float64(ysteps)
	plots := make([][]float64, len(fs))
	for index, xf := range fs {
		plot := make([]float64, xsteps)
		plots[index] = plot
		err = DoPlot(plot, xf, xsteps, ysteps, xmin, xmax, ymin, ymax, xstep)
		if err != nil {
			return
		}
	}
	PrintPlot(xsteps, ysteps, ymin, ymax, ystep, plots)
	return
}

// Plot the values of the supplied function.
func DoPlot(plot []float64, f PlotFunc, xsteps, ysteps int,
	xmin, xmax, ymin, ymax, xstep float64) (err error) {
	xvalue := xmin
	for i := 0; i < xsteps; i++ {
		v := f(xvalue)
		if v < ymin || v > ymax {
			err = ErrBadRange
			return
		}
		xvalue += xstep
		plot[i] = v
	}
	return
}

// Print the plots of the supplied data.
func PrintPlot(xsteps, ysteps int, ymin float64, ymax float64, ystep float64,
	plots [][]float64) {
	if xsteps <= 0 || ysteps <= 0 {
		return
	}
	middle := ysteps / 2
	for yIndex := 0; yIndex < ysteps; yIndex++ {
		fmt.Printf("%8.2f: ", math.Round((ymax-float64(yIndex)*ystep)*100)/100)
		ytop, ybottom := ymax-float64(yIndex)*ystep, ymax-float64(yIndex+1)*ystep
		for xIndex := 0; xIndex < xsteps; xIndex++ {
			pv := " "
			if yIndex == middle {
				pv = "-"
			}
			for plotIndex := 0; plotIndex < len(plots); plotIndex++ {
				v := plots[plotIndex][xIndex]
				if v <= ytop && v >= ybottom {
					pv = string(markers[plotIndex%len(markers)])
				}
			}
			fmt.Print(pv)
		}
		fmt.Println()
	}
	fmt.Printf("%8.2f: ", math.Round((ymax-float64(ysteps+1)*ystep)*100)/100)
}

const markers = "*.^~-=+"

func testPlotPrint() {
	err := PlotPrinter(100, 20, 0, 4*math.Pi, -1.5, 4,
		func(in float64) float64 {
			return math.Sin(in)
		}, func(in float64) float64 {
			return math.Cos(in)
		}, func(in float64) float64 {
			if in == 0 {
				return 0
			}
			return math.Sqrt(in) / in
		})
	if err != nil {
		fmt.Printf("plotting failed: %v", err)
	}
}


func main() {
	testPlotPrint()
}

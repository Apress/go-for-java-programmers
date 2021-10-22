package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
)


// Command line flags.
var (
	urlFlag         string
	nameFlag        string
	gridFlag        string
	magFactorFlag   int
	startServerFlag bool
	runTimingsFlag  bool
	reportFlag      bool
	saveImageFlag   bool
)

// Command line help strings
const (
	urlHelp       = "URL of the PNG image to load"
	nameHelp      = "name to refer to the game initialized by the URL"
	magFactorHelp = "magnify the grid by this factor when formatted into an image"
	gridHelp      = "specify the layout grid (for PNG images); MxN, default 1x1"
	startHelp     = "start the HTTP server (default true)"
	timingHelp    = "run game cycle timings with different goroutine counts"
	reportHelp    = "output run statistics"
	saveImageHelp = "save generated images into a file"
)

// Define command line flags.
// Some are aliases (short forms).
func init() {
	flag.StringVar(&urlFlag, "url", "", urlHelp)
	flag.StringVar(&urlFlag, "u", "", urlHelp)
	flag.StringVar(&nameFlag, "name", "", nameHelp)
	flag.StringVar(&nameFlag, "n", "", nameHelp)
	flag.StringVar(&gridFlag, "grid", "1x1", gridHelp)
	flag.IntVar(&magFactorFlag, "magFactor", 1, magFactorHelp)
	flag.IntVar(&magFactorFlag, "mf", 1, magFactorHelp)
	flag.IntVar(&magFactorFlag, "mag", 1, magFactorHelp)
	flag.BoolVar(&startServerFlag, "start", true, startHelp)
	flag.BoolVar(&runTimingsFlag, "time", false, timingHelp)
	flag.BoolVar(&reportFlag, "report", false, reportHelp)
	flag.BoolVar(&saveImageFlag, "saveImage", false, saveImageHelp)
	flag.BoolVar(&saveImageFlag, "si", false, saveImageHelp)
}

const golDescription = `
Play the game of Life.
Game boards are initialized from PNG images.
Games play over cycles.
Optionally acts as a server to retrieve images of game boards during play.
No supported positional arguments. Supported flags (some have short forms):
`

// Main entry point.
// Sample: -n bart -u file:/Users/Administrator/Downloads/bart.png
func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintln(os.Stderr, strings.TrimSpace(golDescription))
		flag.PrintDefaults()
		os.Exit(0)
	}
	fmt.Printf("Command arguments: %v\n", os.Args[1:])
	fmt.Printf("Go version: %v\n", runtime.Version())
	flag.Parse() // parse any flags
	if len(flag.Args()) > 0 {
		fatalIfError(fmt.Fprintf(os.Stderr,
			"positional command arguments (%v) not accepted\n", flag.Args()))
		os.Exit(1)
	}
	launch()
}

func launch() {
	if len(urlFlag) > 0 {
		if len(nameFlag) == 0 {
			fatalIfError(fmt.Fprintln(os.Stderr,
				"a name is required when a URL is provided"))
		}
		if runTimingsFlag {
			runCycleTimings()
		}
	}

	if startServerFlag {
		startHTTPServer()
	}
}

// launch the HTTP server.
func startHTTPServer() {
	err := startServer()
	if err != nil {
		fmt.Printf("start Server failed: %v\n", err)
		os.Exit(3)
	}

}

// Output information about recorded cycles.
func runCycleTimings() {
	cpuCount := runtime.NumCPU()
	for i := 1; i <= 64; i *= 2 {
		fmt.Printf("Running with %d goroutines, %d CPUs...\n", i, cpuCount)
		CoreGame.GoroutineCount = i
		err := CoreGame.Run(nameFlag, urlFlag)
		if err != nil {
			fmt.Printf("Program failed: %v\n", err)
			os.Exit(2)
		}
		if reportFlag {
			fmt.Printf("Game max: %d, go count: %d:\n",
				CoreGame.MaxCycles, CoreGame.GoroutineCount)
			for _, gr := range CoreGame.Runs {
				fmt.Printf("Game Run: %v, cycle count: %d\n", gr.Name, len(gr.Cycles))
				for _, c := range gr.Cycles {
					start, end :=
						c.StartedAt.UnixNano()/NanosPerMs,
						c.EndedAt.UnixNano()/NanosPerMs
					fmt.Printf(
						"Cycle: start epoch: %dms, end epoch: %dms, elapsed: %dms\n",
						start, end, end-start)
				}
			}
		}
	}
}

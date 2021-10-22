package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image/gif"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var spec = ":8080" // means localhost:8080

// launch HTTP server for th GoL.
func startServer() (err error) {
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/show", showHandler)
	http.HandleFunc("/history", historyHandler)
	fmt.Printf("Starting Server %v...\n", spec)
	err = http.ListenAndServe(spec, nil)
	return
}

// XYyyy types are returned to clients as JSON or XML.
// They are subset of Yyyy types used by the game player.
// They have no reference loops (i.e., to parents) not allowed in JSON and
// omit large fields.
// The tags define how the data is named and formatted

// Represents a game.
type XGame struct {
	Runs map[string]*XGameRun
}

type XGameCycle struct {
	Cycle           int   `json:"cycle" xml:"Cycle"`
	StartedAt       int64 `json:"startedAtNS" xml:"StartedAtEpochNS"`
	EndedAt         int64 `json:"endedAtNS" xml:"EndedAtEpochNS"`
	Duration        int64 `json:"durationMS" xml:"DurationMS"`
	GorountineCount int   `json:"goroutineCount" xml:"GorountineCount"`
	MaxCycles       int   `json:"maximumCycles" xml:"MaximumCycles"`
}

type XGameRun struct {
	Name        string        `json:"name" xml:"Name"`
	ImageURL    string        `json:"imageURL" xml:"ImageURL"`
	StartedAt   int64         `json:"startedAtNS" xml:"StartedAtEpochNS"`
	EndedAt     int64         `json:"endedAtNS" xml:"EndedAtEpochNS"`
	Duration    int64         `json:"durationMS" xml:"DurationMS"`
	Width       int           `json:"width" xml:"Width"`
	Height      int           `json:"height" xml:"Height"`
	Cycles      []*XGameCycle `json:"gameCycles" xml:"GameCycles>GameCycle,omitempty"`
	DelayIn10ms int           `json:"delay10MS" xml:"Delay10MS"`
	PlayIndex   int           `json:"playIndex" xml:"PlayIndex"`
}

func getLead(s string) (res string) {
	res = s
	posn := strings.Index(s, "?")
	if posn >= 0 {
		res = s[0:posn]
	}
	return
}

// History request handler
func historyHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		if getLead(request.RequestURI) != "/history" {
			writer.WriteHeader(405)
			return
		}
		game := &XGame{}
		game.Runs = make(map[string]*XGameRun)
		for k, g := range CoreGame.Runs {
			game.Runs[k] = makeReturnedRun(k, g.ImageURL)
		}
		ba, err := json.MarshalIndent(game, "", "  ")
		if err != nil {
			writer.WriteHeader(500)
			return
		}
		writer.Header().Add("Content-Type", "text/json")
		writer.WriteHeader(200)
		writer.Write(ba) // send response; error ignored
	case "DELETE":
		if request.RequestURI != "/history" {
			writer.WriteHeader(405)
			return
		}
		for k, _ := range CoreGame.Runs {
			delete(CoreGame.Runs, k)
		}
		writer.WriteHeader(204)
	default:
		writer.WriteHeader(405)
	}
}

// Play request handler.
func playHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" || getLead(request.RequestURI) != "/play" {
		writer.WriteHeader(405)
		return
	}
	err := request.ParseForm() // get query parameters
	if err != nil {
		writer.WriteHeader(400)
		return
	}
	name := request.Form.Get("name")
	url := request.Form.Get("url")
	if len(url) == 0 || len(name) == 0 {
		writer.WriteHeader(400)
		return
	}
	ct := request.Form.Get("ct")
	if len(ct) == 0 {
		ct = request.Header.Get("content-type")
	}
	ct = strings.ToLower(ct)
	switch ct {
	case "":
		ct = "application/json"
	case "application/json", "text/json":
	case "application/xml", "text/xml":
	default:
		writer.WriteHeader(400)
		return
	}

	err = CoreGame.Run(name, url)
	if err != nil {
		writer.WriteHeader(500)
		return
	}
	run := makeReturnedRun(name, url)

	var ba []byte
	switch ct {
	case "application/json", "text/json":
		ba, err = json.MarshalIndent(run, "", "  ")
		if err != nil {
			writer.WriteHeader(500)
			return
		}
		writer.Header().Add("Content-Type", "text/json")
	case "application/xml", "text/xml":
		ba, err = xml.MarshalIndent(run, "", "  ")
		if err != nil {
			writer.WriteHeader(500)
			return
		}
		writer.Header().Add("Content-Type", "text/xml")
	}

	writer.WriteHeader(200)
	writer.Write(ba) // send response; error ignored
}

// Build data for returned run.
func makeReturnedRun(name, url string) *XGameRun {
	run := CoreGame.Runs[name]
	xrun := &XGameRun{}
	xrun.Name = run.Name
	xrun.ImageURL = url
	xrun.PlayIndex = run.PlayIndex
	xrun.DelayIn10ms = run.DelayIn10ms
	xrun.Height = run.Height
	xrun.Width = run.Width
	xrun.StartedAt = run.StartedAt.UnixNano()
	xrun.EndedAt = run.EndedAt.UnixNano()
	xrun.Duration = (xrun.EndedAt - xrun.StartedAt + NanosPerMs/2) / NanosPerMs
	xrun.Cycles = make([]*XGameCycle, 0, 100)

	for _, r := range run.Cycles {
		xc := &XGameCycle{}
		xc.StartedAt = r.StartedAt.UnixNano()
		xc.EndedAt = r.EndedAt.UnixNano()
		xc.Duration = (xc.EndedAt - xc.StartedAt + NanosPerMs/2) / NanosPerMs
		xc.Cycle = r.Cycle
		xc.GorountineCount = CoreGame.GoroutineCount
		xc.MaxCycles = CoreGame.MaxCycles
		xrun.Cycles = append(xrun.Cycles, xc)
	}
	return xrun
}

var re = regexp.MustCompile(`^(\d+)x(\d+)$`)

// Show request handler.
func showHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" || getLead(request.RequestURI) != "/show" {
		writer.WriteHeader(405)
		return
	}
	err := request.ParseForm() // get query parameters
	if err != nil {
		writer.WriteHeader(400)
		return
	}
	name := request.Form.Get("name")
	if len(name) == 0 {
		name = "default"
	}
	form := request.Form.Get("form")
	if len(form) == 0 {
		form = "gif"
	}
	xmaxCount := request.Form.Get("maxCount")
	if len(xmaxCount) == 0 {
		xmaxCount = "20"
	}
	maxCount, err := strconv.Atoi(xmaxCount)
	if err != nil || maxCount < 1 || maxCount > 100 {
		writer.WriteHeader(400)
		return
	}
	xmag := request.Form.Get("mag")
	if len(xmag) > 0 {
		mag, err := strconv.Atoi(xmag)
		if err != nil || mag < 1 || mag > 20 {
			writer.WriteHeader(400)
			return
		}
		magFactorFlag = mag
	}

	index := 0
	// verify parameters based on type
	switch form {
	case "gif", "GIF":
	case "png", "PNG":
		xindex := request.Form.Get("index")
		if len(xindex) == 0 {
			xindex = "0"
		}
		index, err = strconv.Atoi(xindex)
		if err != nil {
			writer.WriteHeader(400)
			return
		}
		xgrid := request.Form.Get("grid")
		if len(xgrid) > 0 {
			parts := re.FindStringSubmatch(xgrid)
			if len(parts) != 2 {
				writer.WriteHeader(400)
				return
			}
			gridFlag = fmt.Sprintf("%sx%s", parts[0], parts[1])
		}
	default:
		writer.WriteHeader(400)
		return
	}

	gr, ok := CoreGame.Runs[name]
	if ! ok {
		writer.WriteHeader(404)
		return
	}
	// return requested image type
	switch form {
	case "gif", "GIF":
		gifs, err := gr.MakeGIFs(maxCount)
		if err != nil {
			writer.WriteHeader(500)
			return
		}
		var buf bytes.Buffer
		err = gif.EncodeAll(&buf, gifs)
		if err != nil {
			writer.WriteHeader(500)
			return
		}
		count, err := writer.Write(buf.Bytes()) // send response
		log.Printf("Returned GIF, size=%d\n", count)
		if saveImageFlag {
			saveFile := fmt.Sprintf("/temp/Image_%s.gif", name)
			xerr := ioutil.WriteFile(saveFile, buf.Bytes(), os.ModePerm)
			fmt.Printf("Save %s: %v\n", saveFile, xerr)
		}
	case "png", "PNG":
		if gridFlag == "1x1" {
			if index <= maxCount {
				var buf bytes.Buffer
				err = gr.MakePNG(&buf, index)
				if err != nil {
					code := 500
					if err == BadIndexError {
						code = 400
					}
					writer.WriteHeader(code)
					return
				}
				writer.Write(buf.Bytes()) // send response; error ignored
			} else {
				writer.WriteHeader(400)
			}
		} else {
			// currently not implemented
			writer.WriteHeader(400)
		}
	}
}
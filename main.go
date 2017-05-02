package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

var streams []stream
var commands []*exec.Cmd

type stream struct {
	Name   string
	Stream string
	Image  string
}

func main() {
	//Read in urls of webcams from configuration file
	data, err := ioutil.ReadFile("streams.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &streams); err != nil {
		panic(err)
	}

	//Startup webserver to listen to commands to execute video player
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/all", serveAll)
	http.HandleFunc("/pick", serveOne)
	http.HandleFunc("/stop", serveStop)
	// showAll()
	log.Println("server started...")
	log.Fatal(http.ListenAndServe(":2000", nil))
}

func renderWebsite(w http.ResponseWriter) {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Zoo Cam Viewer</title>
	</head>
	<body>
    <div><h1><a href="/all">All</a></h1></div>
		<div><h1><a href="/stop">Stop</a></h1></div>
		{{range .Streams}}<div>{{ .Name }}</div><div><a href="/pick?name={{.Name}}"><img src="{{.Image}}"/></a></div>{{else}}<div><strong>no streams</strong></div>{{end}}
	</body>
</html>`
	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		panic(err)
	}
	t.Execute(w, struct{ Streams []stream }{Streams: streams})
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	//Serve up basic website with icons to choose from with "all" option
	renderWebsite(w)
}

func serveAll(w http.ResponseWriter, r *http.Request) {
	//Close all existing processes, and fire up all the videos
	showAll()
	renderWebsite(w)
}

func serveOne(w http.ResponseWriter, r *http.Request) {
	//Close all existing processes, and fire up the single video passed in
	name := r.FormValue("name")
	log.Println("got stream name of " + name)
	for _, stream := range streams {
		if stream.Name == name {
			log.Println("Starting stream " + stream.Name)
			showOne(stream)
		}
	}
	renderWebsite(w)
}

func serveStop(w http.ResponseWriter, r *http.Request) {
	killAll()
	renderWebsite(w)
}

func showAll() {
	killAll()
	width := 1920
	height := 1080
	streamCount := len(streams)

	//Determine how many streams we have to make even boxed grids
	boxes := 1
	for ; boxes*boxes < streamCount; boxes++ {
	}

	startWidth := 0
	startHeight := 0
	widthStep := width / boxes
	heightStep := height / boxes
	//We now have a box X box width screen (say 3x3), so split the screen appropriately
	for index, s := range streams {
		endWidth := startWidth + ((index + 1) * widthStep)
		endHeight := startHeight + ((index + 1) * heightStep)
		log.Printf("end width is %v and end height is %v\n", endWidth, endHeight)
		log.Printf("dimensions of window: %v,%v,%v,%v", startWidth, startHeight, endWidth, endHeight)
		cmd := exec.Command("omxplayer", "--win", fmt.Sprintf("%v,%v,%v,%v", startWidth, startHeight, endWidth, endHeight), s.Stream)
		// cmd := exec.Command("mplayer", s.Stream)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		commands = append(commands, cmd)
		cmd.Start()
		time.Sleep(5 * time.Second)
	}
}

func showOne(s stream) {
	killAll()
	//Startup in fullscreen
	cmd := exec.Command("omxplayer", "-b", s.Stream)
	// cmd := exec.Command("mplayer", s.Stream)
	cmd.Start()
	commands = append(commands, cmd)
}

func killAll() {
	log.Println("killing all existing streams")
	for _, proc := range commands {
		proc.Process.Kill()
	}
}

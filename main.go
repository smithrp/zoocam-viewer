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
)

var streams []stream
var boxes []screen

// var commands []*exec.Cmd

type stream struct {
	Name     string
	Stream   string
	Image    string
	Favorite bool
}

type screen struct {
	StartX int
	StartY int
	EndX   int
	EndY   int
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

	setupBoxes()

	//Startup webserver to listen to commands to execute video player
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/all", serveAll)
	http.HandleFunc("/pick", serveOne)
	http.HandleFunc("/stop", serveStop)
	// showAll()
	log.Println("server started...")
	log.Fatal(http.ListenAndServe(":8080", nil))
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

	favoriteStreams := []stream{}
	for _, s := range streams {
		if s.Favorite {
			favoriteStreams = append(favoriteStreams, s)
		}
	}

	for index, s := range favoriteStreams {
		if index >= len(boxes) {
			panic("Boxes were not properly configured")
		}
		box := boxes[index]
		cmd := exec.Command("omxplayer", "--win", fmt.Sprintf("%v,%v,%v,%v", box.StartX, box.StartY, box.EndX, box.EndY), s.Stream)
		// cmd := exec.Command("mplayer", s.Stream)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		// commands = append(commands, cmd)
		cmd.Start()
		// time.Sleep(1 * time.Second)
	}
}

func showOne(s stream) {
	killAll()
	//Startup in fullscreen
	cmd := exec.Command("omxplayer", "-b", s.Stream)
	// cmd := exec.Command("mplayer", s.Stream)
	cmd.Start()
	// commands = append(commands, cmd)
}

func killAll() {
	log.Println("killing all existing streams")
	cmd := exec.Command("sudo", "killall", "omxplayer.bin")
	cmd.Run()
	// for _, proc := range commands {
	// 	// log.Println("killing process ", proc.Process.Pid)
	// 	proc.Process.Kill()
	// }
}

func setupBoxes() {
	//For now we're going to hard code a 3x2 grid across the screen
	width := 1920
	height := 1080

	//Create boxes based on the width and height of the screen

	widthStep := width / 3
	heightStep := height / 2

	//Row1
	for i := 0; i < 3; i++ {
		box := screen{
			StartX: i * widthStep,
			StartY: 0,
			EndX:   (i * widthStep) + widthStep,
			EndY:   heightStep,
		}
		boxes = append(boxes, box)
	}
	//Row2
	for i := 0; i < 3; i++ {
		box := screen{
			StartX: i * widthStep,
			StartY: heightStep,
			EndX:   (i * widthStep) + widthStep,
			EndY:   heightStep + heightStep,
		}
		boxes = append(boxes, box)
	}

	for _, box := range boxes {
		log.Printf("screen size: %v,%v,%v,%v\n", box.StartX, box.StartY, box.EndX, box.EndY)
	}
}

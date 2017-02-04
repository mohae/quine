package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

const mainFile = "main.go"

var (
	exe    = filepath.Base(os.Args[0]) // name of executable
	app    string
	path   string
	cmdDir bool
)

func init() {
	flag.StringVar(&app, "app", "", "name of the application; only use if it is different than the name of the repo")
	flag.StringVar(&path, "path", "", "path of project repo, relative to $GOPATH/src; if empty the WD will be used")
	flag.BoolVar(&cmdDir, "cmd", true, "use a cmd directory for package main")

	log.SetFlags(0)
	log.SetPrefix(exe + ": ")
}

func main() {
	parseFlags()

	// make the output dir, just in case it doesn't exist
	err := os.MkdirAll(path, 0764)
	if err != nil {
		log.Fatalf("error: mkdirall: %s", err)
	}

	// exit using whatever is returned as the return code
	os.Exit(generate())
}

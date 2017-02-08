package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/mohae/linewrap"
)

const mainFile = "main.go"

var (
	exe        = filepath.Base(os.Args[0]) // name of executable
	quinePath  string
	licenseDir = "license"
	license    string

	app App
)

// App is the app that quine is to generate.
type App struct {
	Name string
	Path string
	License
	CmdDir  bool
	buf     bytes.Buffer
	wrapper linewrap.Wrap
	Owner   string // the owner of the copyright.
	Year    string // the year of t he copyright; current year
}

func init() {
	// set app information
	var err error
	app.Owner, err = githubUsername()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: error: %s\n", exe, err)
		fmt.Fprintf(os.Stderr, "%s: the copyright owner information will not be set unless it is either provided via flag or config\n")
	}
	app.Year = strconv.Itoa(time.Now().Year())
	app.wrapper = linewrap.New()
	app.wrapper.LineComment(true)

	quinePath = os.Getenv("QUINEPATH")
	flag.StringVar(&app.Name, "app", "", "name of the application; only use if it is different than the name of the repo")
	flag.StringVar(&license, "license", "", "name of license for the project; use the SPDX short identifier for the language: https://spdx.org/licenses/")
	flag.StringVar(&licenseDir, "licensedir", licenseDir, "the directory that the licenses are in; this is joined with the quinepath or WD to make the full path to the license directory")
	flag.StringVar(&app.Path, "path", "", "path of project repo, relative to $GOPATH/src; if empty the WD will be used")
	flag.StringVar(&quinePath, "quinepath", quinePath, "path for quine application resources; e.g. license")
	flag.StringVar(&app.Owner, "owner", app.Owner, "name of the copyright owner")
	flag.StringVar(&app.Year, "year", app.Year, "yyyy for copyright")
	flag.BoolVar(&app.CmdDir, "cmd", false, "use a cmd directory for package main")

	log.SetFlags(0)
	log.SetPrefix(exe + ": ")
}

func main() {
	parseFlags()

	// make the output dir, just in case it doesn't exist
	err := os.MkdirAll(app.Path, 0764)
	if err != nil {
		log.Fatalf("error: mkdirall: %s", err)
	}

	// exit using whatever is returned as the return code
	os.Exit(app.Generate())
}

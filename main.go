package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"path/filepath"

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
}

func init() {
	quinePath = os.Getenv("QUINEPATH")
	flag.StringVar(&app.Name, "app", "", "name of the application; only use if it is different than the name of the repo")
	flag.StringVar(&license, "license", "", "name of license for the project; use the SPDX short identifier for the language: https://spdx.org/licenses/")
	flag.StringVar(&licenseDir, "licensedir", licenseDir, "the directory that the licenses are in; this is joined with the quinepath or WD to make the full path to the license directory")
	flag.StringVar(&app.Path, "path", "", "path of project repo, relative to $GOPATH/src; if empty the WD will be used")
	flag.StringVar(&quinePath, "quinepath", quinePath, "path for quine application resources; e.g. license")
	flag.BoolVar(&app.CmdDir, "cmd", false, "use a cmd directory for package main")

	log.SetFlags(0)
	log.SetPrefix(exe + ": ")
	app.wrapper = linewrap.New()
	app.wrapper.Indent = true
	app.wrapper.IndentVal = "// "
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

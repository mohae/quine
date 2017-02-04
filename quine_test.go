package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	var err error
	app = "test"
	path, err = ioutil.TempDir("", "quine")
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestWriteOpening(t *testing.T) {
	var buf bytes.Buffer
	expected := `package main

import (
	"flag"
	"os"
	"path/filepath"
)

var app = filepath.Base(os.Args[0]) // name of application
var cfg Config

type Config struct {
	LogFile string   // output destination for logs; stderr is default
	f       *os.File // logfile handle for close; this will be nil if output is stderr
}

func init() {
	flag.StringVar(&cfg.LogDst, "logfile", "stderr", "output destination for logs")
}

func main() {
	// Process flags
	parseFlags()
	os.Exit(testMain())
}
`
	err := writeMain(&buf)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	// check the written file
	b, err := ioutil.ReadFile(filepath.Join(path, mainFile))
	if string(b) != expected {
		t.Errorf("got %s\nwant %s", string(b), expected)
	}
}

func TestWriteAppFile(t *testing.T) {
	var buf bytes.Buffer
	expected := `package main

import (
	"fmt"
	"log"
	"os"
)

// parseFlag handles flag parsing, validation, and any side affects of flag
// states. Errors or invalid states should result in printing a message to
// os.Stderr and an os.Exit() with a non-zero int.
func parseFlag() {
	var err error

	flag.Parse()

	if cfg.LogFile != "" && cfg.LogFile != "stdout" { // open the logfile if one is specified
		cfg.f, err = os.FileOpen(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: open logfile: %s", app, err)
			os.Exit(1)
		}
	}
}

func testMain() int {
	if cfg.f != nil {
		defer f.Close() // make sure the logfile is closed if there is one
	}

	fmt.Printf("%s: hello, world\n", app)

	return 0
}
`
	err := writeAppFile(&buf)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	// check the written file
	b, err := ioutil.ReadFile(filepath.Join(path, app+"_main.go"))
	if string(b) != expected {
		t.Errorf("got %s\nwant %s", string(b), expected)
	}
}

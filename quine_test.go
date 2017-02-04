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
	"path/filepath"
)

var app = filepath.Base(os.Args[0]) // name of application

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
)

// parseFlag handles flag parsing, validation, and any side affects of flag
// states. Errors or invalid states should result in printing a message to
// os.Stderr and an os.Exit() with a non-zero int.
func parseFlag() {
	flag.Parse()
}

func testMain() int {
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

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"

	"github.com/mohae/linewrap"
)

func parseFlags() {
	flag.Parse()
	var err error
	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			log.Fatalf("error: get WD: ", err)
		}
	} else {
		// build it relative to GOPATH
		// TODO windows
		gop := os.Getenv("GOPATH")
		if gop == "" { // if it wasn't set, use Go's default path (1.8) + src
			gop = "$HOME/go/src"
		}
		path = filepath.Join(gop, path)
	}
	// set the app name, if it isn't set
	if app == "" {
		app = filepath.Base(path)
	}

	if cmdDir { // adjust output path if there is going to be a command directory
		path = filepath.Join(path, "cmd", app)
	}
}

// generate does the actual work of creating the main.go and whatever else is
// needed
func generate() int {
	// build everything first in a buffer so it can be fmt'd before writing to file.
	var buf bytes.Buffer

	// these are in separate funcs for testability
	err := writeMain(&buf)
	if err != nil {
		log.Printf("%s: error: %s", mainFile, err)
		return 1
	}

	buf.Reset()

	err = writeAppFile(&buf)
	if err != nil {
		log.Printf("%s: error: %s", app+"_main.go", err)
		return 1
	}

	return 0
}

func writeMain(buf *bytes.Buffer) error {
	_, err := buf.WriteString("package main\nimport (\n\"flag\"\n\"path/filepath\"\n\"os\"\n)\n\nvar app = filepath.Base(os.Args[0]) // name of application\n")
	if err != nil {
		return err
	}

	// config
	_, err = buf.WriteString("var cfg Config\n\ntype Config struct {\nLogFile string // output destination for logs; stderr is default\nf *os.File // logfile handle for close; this will be nil if output is stderr\n}\n")
	if err != nil {
		return err
	}

	// init
	_, err = buf.WriteString("\nfunc init() {\nflag.StringVar(&cfg.LogDst, \"logfile\", \"stderr\", \"output destination for logs\")\n}\n")
	if err != nil {
		return err
	}

	// main
	_, err = buf.WriteString("\nfunc main() {\n// Process flags\nparseFlags()\nos.Exit(")
	if err != nil {
		return err
	}
	_, err = buf.WriteString(app)
	if err != nil {
		return err
	}
	_, err = buf.WriteString("Main())\n}")
	if err != nil {
		return err
	}

	// fmt the code
	fmtd, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("fmt source: %s", err)
	}

	// open the file and write
	f, err := os.OpenFile(filepath.Join(path, mainFile), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0664)
	if err != nil {
		return fmt.Errorf("open failed: %s", err)
	}
	defer f.Close()

	n, err := f.Write(fmtd)
	if err != nil {
		return fmt.Errorf("write failed: %s", err)
	}

	fmt.Printf("%s: %d bytes were written to %s\n", exe, n, filepath.Join(path, mainFile))
	return nil
}

// write the app.go file.
func writeAppFile(buf *bytes.Buffer) error {
	appFile := filepath.Join(path, app+"_main.go")
	// if the app file already exists; don't modify to prevent overwriting any user code.
	_, err := os.Stat(appFile)
	if err == nil {
		return nil
	}
	if err != nil && !os.IsNotExist(err) { // if the error wasn't IsNotExist, return the err.
		return fmt.Errorf("%s: %s", appFile, err)
	}

	_, err = buf.WriteString("package main\nimport(\n\"fmt\"\n\"log\"\n\"os\"\n)\n")
	if err != nil {
		return err
	}

	err = writeParseFlag(buf)
	if err != nil {
		return err
	}

	_, err = buf.WriteString("\n\nfunc ")
	if err != nil {
		return err
	}

	_, err = buf.WriteString(app)
	if err != nil {
		return err
	}

	_, err = buf.WriteString("Main() int {\nif cfg.f != nil {\ndefer f.Close() // make sure the logfile is closed if there is one\n}\n\nfmt.Printf(\"%s: hello, world\\n\", app)\n\nreturn 0\n}\n")
	if err != nil {
		return err
	}

	// fmt the code
	fmtd, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("fmt source: %s", err)
	}

	// open the file and write
	f, err := os.OpenFile(appFile, os.O_CREATE|os.O_RDWR, 0664)
	if err != nil {
		return fmt.Errorf("open failed: %s", err)
	}
	defer f.Close()

	n, err := f.Write(fmtd)
	if err != nil {
		return fmt.Errorf("write failed: %s", err)
	}

	fmt.Printf("%s: %d bytes were written to %s\n", exe, n, appFile)
	return nil
}

// write the parseFlag func: parseFlag os.Exit's on any error.
func writeParseFlag(buf *bytes.Buffer) error {
	lw := linewrap.New()
	lw.Indent = true
	lw.IndentVal = "// "
	cmt := "// parseFlag handles flag parsing, validation, and any side affects of flag states. Errors or invalid states should result in printing a message to os.Stderr and an os.Exit() with a non-zero int."
	cmt, err := lw.Line(cmt)
	if err != nil {
		return fmt.Errorf("parseFlag func: %s", err)
	}

	_, err = buf.WriteString(cmt)
	if err != nil {
		return fmt.Errorf("parseFlag func: %s", err)
	}

	_, err = buf.WriteString("\nfunc parseFlag() {\nvar err error\n\nflag.Parse()\n\n")
	if err != nil {
		return fmt.Errorf("parseFlag func: %s", err)
	}

	// log
	_, err = buf.WriteString("if cfg.LogFile != \"\" && cfg.LogFile != \"stdout\" {  // open the logfile if one is specified\ncfg.f, err = os.FileOpen(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)\n")
	if err != nil {
		return fmt.Errorf("parseFlag func: %s", err)
	}
	_, err = buf.WriteString("if err != nil {\nfmt.Fprintf(os.Stderr, \"%s: open logfile: %s\", app, err)\nos.Exit(1)\n}\n}\n}\n")
	if err != nil {
		return fmt.Errorf("parseFlag func: %s", err)
	}

	return nil

}

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mohae/linewrap"
)

func parseFlags() {
	flag.Parse()
	var err error
	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s error: get WD: ", app, err)
			os.Exit(1)
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

	licenseType, err = LicenseFromString(license)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s error: %s", app, err)
		os.Exit(1)
	}
}

// generate does the actual work of creating the main.go and whatever else is
// needed
func generate() int {
	// If a license was specified, copy it to the path.
	if licenseType != None {
		err := copyLicense()
		if err != nil {
			log.Printf("copy %s: error: %s", licenseType, err)
			return 1
		}

		if err != nil {
			log.Printf("copy %s license: error: %s", licenseType, err)
			return 1
		}
	}

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
	// if a license was specified, open its notice file and write it to main.go
	if licenseType != None {
		noticeFile := filepath.Join(quinePath, licenseDir, licenseType.ID()+".notice")
		f, err := os.Open(noticeFile)
		if err != nil {
			if os.IsNotExist(err) { // not all licenses have notices
				goto writeMain
			}
			return fmt.Errorf("open %s: %s", noticeFile, err) // return any other error
		}
		defer f.Close()
		// TODO do element replacement for the notices that have that.
		_, err = io.Copy(buf, f)
		if err != nil {
			return fmt.Errorf("copy %s: %s", noticeFile, err)
		}
		buf.WriteString("\n\n")
	}

writeMain:
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
	_, err = buf.WriteString("\nfunc init() {\nflag.StringVar(&cfg.LogDst, \"logfile\", \"stderr\", \"output destination for logs\")\n\nlog.SetPrefix(app + \": \")\n}\n")
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

func copyLicense() error {
	lFile := strings.ToLower(licenseType.ID())

	srcFile := filepath.Join(quinePath, licenseDir, lFile)
	src, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("open source file: %s", err)
	}
	defer src.Close()

	dstFile := filepath.Join(path, lFile)
	dst, err := os.OpenFile(dstFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0664)
	if err != nil {
		return fmt.Errorf("open dest. file: %s", err)
	}
	defer dst.Close()

	n, err := io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("copy license from %s to %s: %s", srcFile, dstFile, err)
	}

	fmt.Printf("%s copied to %s; %d bytes written\n", app, lFile, dstFile, n)
	return nil

}

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func parseFlags() {
	flag.Parse()
	var err error
	if app.Path == "" {
		app.Path, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s error: get WD: ", app.Name, err)
			os.Exit(1)
		}
	} else {
		// build it relative to GOPATH
		// TODO windows
		gop := os.Getenv("GOPATH")
		if gop == "" { // if it wasn't set, use Go's default path (1.8) + src
			gop = "$HOME/go/src"
		} else {
			gop = filepath.Join(gop, "src")
		}
		app.Path = filepath.Join(gop, app.Path)
	}
	// set the app name, if it isn't set
	if app.Name == "" {
		app.Name = filepath.Base(app.Path)
	}

	if app.CmdDir { // adjust output path if there is going to be a command directory
		app.Path = filepath.Join(app.Path, "cmd", app.Name)
	}

	app.License, err = LicenseFromString(license)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s error: %s", app.Name, err)
		os.Exit(1)
	}
}

// generate does the actual work of creating the main.go and whatever else is
// needed
func (a *App) Generate() int {
	// If a license was specified, copy it to the path.
	if a.License != None {
		err := a.CopyLicense()
		if err != nil {
			log.Printf("copy %s: error: %s", a.License, err)
			return 1
		}
	}

	// these are in separate funcs for testability
	err := a.WriteMain()
	if err != nil {
		log.Printf("%s: error: %s", mainFile, err)
		return 1
	}

	err = a.WriteAppFile()
	if err != nil {
		log.Printf("%s: error: %s", a.Name+"_main.go", err)
		return 1
	}

	return 0
}

func (a *App) WriteMain() error {
	a.buf.Reset()

	err := a.writeSLH() // write the Standard License Header, if there is one.
	if err != nil {
		return err
	}

	_, err = a.buf.WriteString("package main\nimport (\n\"flag\"\n\"log\"\n\"path/filepath\"\n\"os\"\n)\n\nvar app = filepath.Base(os.Args[0]) // name of application\n")
	if err != nil {
		return err
	}

	// config
	_, err = a.buf.WriteString("var cfg Config\n\ntype Config struct {\nLogFile string // output destination for logs; stderr is default\nf *os.File // logfile handle for close; this will be nil if output is stderr\n}\n")
	if err != nil {
		return err
	}

	// init
	_, err = a.buf.WriteString("\nfunc init() {\nflag.StringVar(&cfg.LogFile, \"logfile\", \"stderr\", \"output destination for logs\")\n\nlog.SetPrefix(app + \": \")\n}\n")
	if err != nil {
		return err
	}

	// main
	_, err = a.buf.WriteString("\nfunc main() {\nflag.Usage = usage\n\n// Process flags\nFlagParse()\n\nos.Exit(")
	if err != nil {
		return err
	}
	_, err = a.buf.WriteString(a.Name)
	if err != nil {
		return err
	}
	_, err = a.buf.WriteString("Main())\n}")
	if err != nil {
		return err
	}

	// fmt the code
	fmtd, err := format.Source(a.buf.Bytes())
	if err != nil {
		return fmt.Errorf("fmt source: %s", err)
	}

	// open the file and write
	f, err := os.OpenFile(filepath.Join(a.Path, mainFile), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0664)
	if err != nil {
		return fmt.Errorf("open failed: %s", err)
	}
	defer f.Close()

	n, err := f.Write(fmtd)
	if err != nil {
		return fmt.Errorf("write failed: %s", err)
	}

	fmt.Printf("%s: %d bytes were written to %s\n", exe, n, filepath.Join(a.Path, mainFile))
	return nil
}

// write the app.go file.
func (a *App) WriteAppFile() error {
	a.buf.Reset()

	appFile := filepath.Join(a.Path, a.Name+"_main.go")
	// if the app file already exists; don't modify to prevent overwriting any user code.
	_, err := os.Stat(appFile)
	if err == nil {
		return nil
	}
	if err != nil && !os.IsNotExist(err) { // if the error wasn't IsNotExist, return the err.
		return fmt.Errorf("%s: %s", appFile, err)
	}

	_, err = a.buf.WriteString("package main\n\nimport(\n\"flag\"\n\"fmt\"\n\"os\"\n)\n")
	if err != nil {
		return err
	}

	err = a.WriteUsage()
	if err != nil {
		return err
	}

	err = a.WriteFlagParse()
	if err != nil {
		return err
	}

	_, err = a.buf.WriteString("\n\nfunc ")
	if err != nil {
		return err
	}

	_, err = a.buf.WriteString(a.Name)
	if err != nil {
		return err
	}

	_, err = a.buf.WriteString("Main() int {\nif cfg.f != nil {\ndefer cfg.f.Close() // make sure the logfile is closed if there is one\n}\n\nfmt.Printf(\"%s: hello, world\\n\", app)\n\nreturn 0\n}\n")
	if err != nil {
		return err
	}

	// fmt the code
	fmtd, err := format.Source(a.buf.Bytes())
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

// write the usage func
func (a *App) WriteUsage() error {
	_, err := a.buf.WriteString("\n")
	if err != nil {
		return fmt.Errorf("usage func: %s", err)
	}

	cmt := "usage is the usage func for flag.Usage."
	cmt, err = a.wrapper.Line(cmt)
	if err != nil {
		return fmt.Errorf("usage func: %s", err)
	}

	_, err = a.buf.WriteString(cmt)
	if err != nil {
		return fmt.Errorf("usage func: %s", err)
	}

	_, err = a.buf.WriteString("\nfunc usage() {\n")
	if err != nil {
		return fmt.Errorf("usage func: %s", err)
	}

	_, err = a.buf.WriteString("fmt.Fprint(os.Stderr, \"Usage:\\n\")\nfmt.Fprintf(os.Stderr, \"  %s [FLAGS] \\n\", app)\nfmt.Fprint(os.Stderr, \"\\n\")\n")
	if err != nil {
		return fmt.Errorf("usage func: %s", err)
	}

	_, err = a.buf.WriteString("fmt.Fprintf(os.Stderr, \"Insert information about %s here\\n\", app)\nfmt.Fprint(os.Stderr, \"\\n\")\nfmt.Fprint(os.Stderr, \"Options:\\n\")\nflag.PrintDefaults()\n")
	if err != nil {
		return fmt.Errorf("usage func: %s", err)
	}

	_, err = a.buf.WriteString("}\n\n")
	if err != nil {
		return fmt.Errorf("usage func: %s", err)
	}
	return nil
}

// write the FlagParse func: parseFlag os.Exit's on any error.
func (a *App) WriteFlagParse() error {
	cmt := "FlagParse handles flag parsing, validation, and any side affects of flag states. Errors or invalid states should result in printing a message to os.Stderr and an os.Exit() with a non-zero int."
	cmt, err := a.wrapper.Line(cmt)
	if err != nil {
		return fmt.Errorf("FlagParse func: %s", err)
	}

	_, err = a.buf.WriteString(cmt)
	if err != nil {
		return fmt.Errorf("FlagParse func: %s", err)
	}

	_, err = a.buf.WriteString("\nfunc FlagParse() {\nvar err error\n\nflag.Parse()\n\n")
	if err != nil {
		return fmt.Errorf("FlagParse func: %s", err)
	}

	// log
	_, err = a.buf.WriteString("if cfg.LogFile != \"\" && cfg.LogFile != \"stdout\" {  // open the logfile if one is specified\ncfg.f, err = os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)\n")
	if err != nil {
		return fmt.Errorf("FlagParse func: %s", err)
	}
	_, err = a.buf.WriteString("if err != nil {\nfmt.Fprintf(os.Stderr, \"%s: open logfile: %s\", app, err)\nos.Exit(1)\n}\n}\n}\n")
	if err != nil {
		return fmt.Errorf("FlagParse func: %s", err)
	}

	return nil

}

// CopyLicense copies the license text. Any placeholders in the text are
// replaced with the actual value; if applicable.
func (a *App) CopyLicense() error {
	lFile := strings.ToLower(a.License.ID())

	srcFile := filepath.Join(quinePath, licenseDir, lFile)
	b, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return fmt.Errorf("reade license file: %s", err)
	}

	// if the license has any placeholders replace them with values
	b = a.replaceLicensePlaceholders(b)
	dstFile := filepath.Join(a.Path, "LICENSE")
	dst, err := os.OpenFile(dstFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0664)
	if err != nil {
		return fmt.Errorf("open dest. file: %s", err)
	}
	defer dst.Close()

	n, err := dst.Write(b)
	if err != nil {
		return fmt.Errorf("write license to %s: %s", dstFile, err)
	}

	fmt.Printf("%s: %s copied to %s; %d bytes written\n", app, lFile, dstFile, n)
	return nil
}

// not all licenses have placeholders to replace.
func (a *App) replaceLicensePlaceholders(b []byte) []byte {
	switch a.License {
	case BSD2Clause:
		return a.replaceBSD2ClauseLicensePlaceholders(b)
	case BSD3Clause:
		return a.replaceBSD3ClauseLicensePlaceholders(b)
	case MIT:
		return a.replaceMITLicensePlaceholders(b)
	}

	return b
}

func (a *App) replaceBSD2ClauseLicensePlaceholders(b []byte) []byte {
	// if owner and year aren't set, nothing to do.
	if a.Owner == "" && a.Year == "" {
		return b
	}

	// make out == the len of the license when replacements are done
	y, o := 6, 7 // <year> <owner>
	if a.Year != "" {
		y = len(a.Year)
	}
	if a.Owner != "" {
		o = len(a.Owner)
	}

	out := make([]byte, 0, len(b)-13+y+o)
	out = append(out, b[:14]...)
	// year
	if a.Year == "" {
		out = append(out, b[14:20]...)
	} else {
		out = append(out, []byte(a.Year)...)
	}

	out = append(out, ' ')

	// owner
	if a.Owner == "" {
		out = append(out, b[21:28]...)
	} else {
		out = append(out, []byte(a.Owner)...)
	}

	out = append(out, b[28:]...)

	return out
}

func (a *App) replaceBSD3ClauseLicensePlaceholders(b []byte) []byte {
	// if owner and year aren't set, nothing to do.
	if a.Owner == "" && a.Year == "" {
		return b
	}

	// make out == the len of the license when replacements are done
	y, o := 6, 7 // <year> <owner>
	if a.Year != "" {
		y = len(a.Year)
	}
	if a.Owner != "" {
		o = len(a.Owner)
	}

	out := make([]byte, 0, len(b)-13+y+o)
	out = append(out, b[:14]...)
	// year
	if a.Year == "" {
		out = append(out, b[14:20]...)
	} else {
		out = append(out, []byte(a.Year)...)
	}

	out = append(out, ' ')

	// owner
	if a.Owner == "" {
		out = append(out, b[21:29]...)
	} else {
		out = append(out, []byte(a.Owner)...)
	}

	out = append(out, b[29:]...)

	return out
}

func (a *App) replaceMITLicensePlaceholders(b []byte) []byte {
	// if owner and year aren't set, nothing to do.
	if a.Owner == "" && a.Year == "" {
		return b
	}

	// make out == the len of the license when replacements are done
	y, o := 6, 19 // <year> <owner>
	if a.Year != "" {
		y = len(a.Year)
	}
	if a.Owner != "" {
		o = len(a.Owner)
	}

	out := make([]byte, 0, len(b)-25+y+o)

	out = append(out, b[:26]...)
	// year
	if a.Year == "" {
		out = append(out, b[26:32]...)
	} else {
		out = append(out, []byte(a.Year)...)
	}

	out = append(out, ' ')
	// owner
	if a.Owner == "" {
		out = append(out, b[33:52]...)
	} else {
		out = append(out, []byte(a.Owner)...)
	}

	out = append(out, b[53:]...)

	return out
}

// If a license was specified, open its SLH, Standard License Header, file, if
// it has one and write it to main.go.
func (a *App) writeSLH() error {
	if a.License == None { // if no license is specified nothing to do
		return nil
	}

	// read the slh file
	slhFile := filepath.Join(quinePath, licenseDir, strings.ToLower(a.License.ID())+".slh")
	b, err := ioutil.ReadFile(slhFile)
	if err != nil {
		if os.IsNotExist(err) { // not all licenses have SLHs, this is not an error state
			return nil
		}
		return fmt.Errorf("SLH file: read %s: %s", slhFile, err) // return any other error
	}

	// MPl-2.0 is used as is.
	switch a.License {
	case Apache20:
		b = a.replaceApache20SLHPlaceholders(b)
	case GPL20:
		b = a.replaceGPL20SLHPlaceholders(b)
	case GPL30:
		b = a.replaceGPL30SLHPlaceholders(b)
	case LGPL20, LGPL21: // format and location of the placeholders are the same for both
		b = a.replaceLGPL2SLHPlaceholders(b)
	}

	cmt, err := a.wrapper.Line(string(b))
	if err != nil {
		return fmt.Errorf("SLH file: format as comment: %s", err)
	}

	_, err = a.buf.WriteString(cmt)
	if err != nil {
		return fmt.Errorf("SLH file: write: %s", err)
	}
	_, err = a.buf.WriteString("\n\n")
	if err != nil {
		return fmt.Errorf("SLH file: write \n: %s", err)
	}
	return nil
}

func (a *App) replaceApache20SLHPlaceholders(b []byte) []byte {
	// if owner and year aren't set, nothing to do.
	if a.Owner == "" && a.Year == "" {
		return b
	}

	// make out == the len of the license when replacements are done
	y, o := 6, 26 // <year> <owner>
	if a.Year != "" {
		y = len(a.Year)
	}
	if a.Owner != "" {
		o = len(a.Owner)
	}

	out := make([]byte, 0, len(b)-33+y+o)
	out = append(out, b[:10]...)
	// year
	if a.Year == "" {
		out = append(out, b[10:16]...)
	} else {
		out = append(out, []byte(a.Year)...)
	}

	out = append(out, ' ')

	// owner
	if a.Owner == "" {
		out = append(out, b[17:42]...)
	} else {
		out = append(out, []byte(a.Owner)...)
	}

	out = append(out, b[42:]...)

	return out
}

func (a *App) replaceGPL20SLHPlaceholders(b []byte) []byte {
	// if owner and year aren't set, nothing to do.
	if a.Owner == "" && a.Year == "" {
		return b
	}

	// make out == the len of the license when replacements are done
	y, o := 4, 14 // yyy name of author
	if a.Year != "" {
		y = len(a.Year)
	}
	if a.Owner != "" {
		o = len(a.Owner)
	}

	out := make([]byte, 0, len(b)-19+y+o)
	out = append(out, b[:14]...)
	// year
	if a.Year == "" {
		out = append(out, b[14:18]...)
	} else {
		out = append(out, []byte(a.Year)...)
	}

	out = append(out, ' ')

	// owner
	if a.Owner == "" {
		out = append(out, b[19:33]...)
	} else {
		out = append(out, []byte(a.Owner)...)
	}

	out = append(out, b[33:]...)

	return out
}

func (a *App) replaceGPL30SLHPlaceholders(b []byte) []byte {
	// if owner and year aren't set, nothing to do.
	if a.Owner == "" && a.Year == "" {
		return b
	}

	// make out == the len of the license when replacements are done
	y, o := 6, 16 // yyy name of author
	if a.Year != "" {
		y = len(a.Year)
	}
	if a.Owner != "" {
		o = len(a.Owner)
	}

	out := make([]byte, 0, len(b)-23+y+o)
	out = append(out, b[:14]...)
	// year
	if a.Year == "" {
		out = append(out, b[14:20]...)
	} else {
		out = append(out, []byte(a.Year)...)
	}

	out = append(out, ' ')

	// owner
	if a.Owner == "" {
		out = append(out, b[21:37]...)
	} else {
		out = append(out, []byte(a.Owner)...)
	}

	out = append(out, b[37:]...)

	return out
}

func (a *App) replaceLGPL2SLHPlaceholders(b []byte) []byte {
	// if owner and year aren't set, nothing to do.
	if a.Owner == "" && a.Year == "" {
		return b
	}

	// make out == the len of the license when replacements are done
	y, o := 4, 14 // yyy name of author
	if a.Year != "" {
		y = len(a.Year)
	}
	if a.Owner != "" {
		o = len(a.Owner)
	}

	out := make([]byte, 0, len(b)-19+y+o)
	out = append(out, b[:14]...)
	// year
	if a.Year == "" {
		out = append(out, b[14:18]...)
	} else {
		out = append(out, []byte(a.Year)...)
	}

	out = append(out, ' ')

	// owner
	if a.Owner == "" {
		out = append(out, b[19:33]...)
	} else {
		out = append(out, []byte(a.Owner)...)
	}

	out = append(out, b[33:]...)

	return out
}

// returns Git's globally configured user.name or an error. This assumes that
// git is installed.
func githubUsername() (gituser string, err error) {
	cmd := exec.Command("git", "config", "--global", "user.name")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("get git global user.name: %s", err)
	}

	return string(bytes.TrimRight(buf.Bytes(), "\n")), nil
}

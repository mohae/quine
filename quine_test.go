package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	var err error
	app.Name = "test"
	quinePath, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

var expectedMain = `package main

import (
	"flag"
	"log"
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
	flag.StringVar(&cfg.LogFile, "logfile", "stderr", "output destination for logs")

	log.SetPrefix(app + ": ")
}

func main() {
	flag.usage = Usage

	// Process flags
	FlagParse()

	os.Exit(testMain())
}
`

func TestWriteMain(t *testing.T) {
	tests := []struct {
		license  License
		expected string
	}{
		//{None, ""},
		{Apache20, `// Copyright [yyyy] [name of copyright owner]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the spec
//

`,
		},
		{BSD3Clause, ""},
		{GPL20, `// <One line to give the program's name and a brief idea of what it does.>
// Copyright (C) <year> <name of author>
//
// This program is free software; you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation; either version 2 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along with
// this program; if not, write to the Free Software Foundation, Inc., 59 Temple
// Place, Suite 330, Boston, MA 02111-1307 USA
//

`,
		},
		{GPL30, `// <one line to give the program's name and a brief idea of what it does.>
// Copyright (C) <year>  <name of author>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

`,
		},
		{LGPL21, `// <one line to give the library's name and an idea of what it does.> Copyright
// (C) <year> <name of author>
//
// This library is free software; you can redistribute it and/or modify it
// under the terms of the GNU Lesser General Public License as published by the
// Free Software Foundation; either version 2.1 of the License, or (at your
// option) any later version.
//
// This library is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License
// for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this library; if not, write to the Free Software Foundation,
// Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307 USA
//

`,
		},
		{LGPL30, ""},
		{MIT, ""},
		{MPL20, `// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.
//

`,
		},
	}
	var err error
	lapp := app
	lapp.Path, err = ioutil.TempDir("", "quine")
	if err != nil {
		panic(err)
	}
	lapp.Owner = "Trillian"
	lapp.Year = "1999"
	for i, test := range tests {
		lapp.License = test.license
		err = lapp.WriteMain()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
			continue
		}
		// check the written file
		b, err := ioutil.ReadFile(filepath.Join(lapp.Path, mainFile))
		if err != nil {
			t.Errorf("unexpected error readging %s: %q", filepath.Join(lapp.Path, mainFile), err)
			continue
		}
		wants := strings.Split(test.expected+expectedMain, "\n")
		gots := strings.Split(string(b), "\n")
		if len(gots) != len(wants) {
			t.Errorf("%d: got %d lines want %d", i, len(gots), len(wants))
			t.Errorf("%d: got %q\nwant %q", i, string(b), test.expected+expectedMain)
			continue
		}
		for j, got := range gots {
			if got != wants[j] {
				t.Errorf("%d:%d: got %q\nwant %q", i, j, got, wants[j])
			}
		}
	}
}

func TestWriteAppFile(t *testing.T) {
	expected := `package main

import (
	"flag"
	"fmt"
	"os"
)

// usage is the usage func for flag.Usage.
func usage() {
	fmt.Fprint(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s [FLAGS] \n", app)
	fmt.Fprint(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Insert information about %s here\n", app)
	fmt.Fprint(os.Stderr, "\n")
	fmt.Fprint(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}

// FlagParse handles flag parsing, validation, and any side affects of flag
// states. Errors or invalid states should result in printing a message to
// os.Stderr and an os.Exit() with a non-zero int.
func FlagParse() {
	var err error

	flag.Parse()

	if cfg.LogFile != "" && cfg.LogFile != "stdout" { // open the logfile if one is specified
		cfg.f, err = os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: open logfile: %s", app, err)
			os.Exit(1)
		}
	}
}

func testMain() int {
	if cfg.f != nil {
		defer cfg.f.Close() // make sure the logfile is closed if there is one
	}

	fmt.Printf("%s: hello, world\n", app)

	return 0
}
`
	var err error
	lapp := app
	lapp.Path, err = ioutil.TempDir("", "quine")
	if err != nil {
		panic(err)
	}
	err = lapp.WriteAppFile()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	// check the written file
	b, err := ioutil.ReadFile(filepath.Join(lapp.Path, lapp.Name+"_main.go"))
	gots := strings.Split(string(b), "\n")
	wants := strings.Split(expected, "\n")
	if len(gots) != len(wants) {
		t.Errorf("got %d lines; want %d", len(gots), len(wants))
		t.Errorf("got %q\nwant %q", string(b), expected)
		//		return
	}
	for i, got := range gots {
		if got != wants[i] {
			t.Errorf("got %q\nwant %q", got, wants[i])
		}
	}
}

func TestReplaceBSD2ClauseLicensePlaceholders(t *testing.T) {
	// only test the first line
	tests := []struct {
		owner    string
		year     string
		expected string
	}{
		{"", "", "Copyright (c) <year> <owner> All rights reserved."},
		{"Zaphod Beeblebrox", "", "Copyright (c) <year> Zaphod Beeblebrox All rights reserved."},
		{"", "1942", "Copyright (c) 1942 <owner> All rights reserved."},
		{"Zaphod Beeblebrox", "1942", "Copyright (c) 1942 Zaphod Beeblebrox All rights reserved."},
	}
	b, err := ioutil.ReadFile(filepath.Join("license", strings.ToLower(BSD2Clause.ID())))
	if err != nil {
		t.Errorf("unexpected error: %q", err)
		return
	}

	a := app
	for i, test := range tests {
		a.Owner = test.owner
		a.Year = test.year
		v := a.replaceBSD2ClauseLicensePlaceholders(b)
		ndx := bytes.IndexByte(v, '\n')
		if ndx < 0 {
			t.Errorf("%d: expected to find a \n; none found", i)
			continue
		}
		line := string(v[:ndx])
		if line != test.expected {
			t.Errorf("%d: got %q want %q", i, line, test.expected)
		}
	}
}

func TestReplaceBSD3ClauseLicensePlaceholders(t *testing.T) {
	// only test the first line
	tests := []struct {
		owner    string
		year     string
		expected string
	}{
		{"", "", "Copyright (c) <year> <owner> . All rights reserved."},
		{"Zaphod Beeblebrox", "", "Copyright (c) <year> Zaphod Beeblebrox. All rights reserved."},
		{"", "1942", "Copyright (c) 1942 <owner> . All rights reserved."},
		{"Zaphod Beeblebrox", "1942", "Copyright (c) 1942 Zaphod Beeblebrox. All rights reserved."},
	}
	b, err := ioutil.ReadFile(filepath.Join("license", strings.ToLower(BSD3Clause.ID())))
	if err != nil {
		t.Errorf("unexpected error: %q", err)
		return
	}

	a := app
	for i, test := range tests {
		a.Owner = test.owner
		a.Year = test.year
		v := a.replaceBSD3ClauseLicensePlaceholders(b)
		ndx := bytes.IndexByte(v, '\n')
		if ndx < 0 {
			t.Errorf("%d: expected to find a \n; none found", i)
			continue
		}
		line := string(v[:ndx])
		if line != test.expected {
			t.Errorf("%d: got %q want %q", i, line, test.expected)
		}
	}
}

func TestReplaceMITLicensePlaceholders(t *testing.T) {
	// only test the first line
	tests := []struct {
		owner    string
		year     string
		expected string
	}{
		{"", "", "Copyright (c) <year> <copyright holders>"},
		{"Zaphod Beeblebrox", "", "Copyright (c) <year> Zaphod Beeblebrox"},
		{"", "1942", "Copyright (c) 1942 <copyright holders>"},
		{"Zaphod Beeblebrox", "1942", "Copyright (c) 1942 Zaphod Beeblebrox"},
	}
	b, err := ioutil.ReadFile(filepath.Join("license", strings.ToLower(MIT.ID())))
	if err != nil {
		t.Errorf("unexpected error: %q", err)
		return
	}

	a := app
	for i, test := range tests {
		a.Owner = test.owner
		a.Year = test.year
		v := a.replaceMITLicensePlaceholders(b)
		start := bytes.IndexByte(v, '\n')
		if start < 0 {
			t.Errorf("%d: expected to find a \n; none found", i)
			continue
		}
		start++
		end := bytes.IndexByte(v[start:], '\n')
		if end < 0 {
			t.Errorf("%d: expected to find a \n; none found", i)
			continue
		}
		end += start

		line := string(v[start:end])
		if line != test.expected {
			t.Errorf("%d: got %q want %q", i, line, test.expected)
		}

	}
}

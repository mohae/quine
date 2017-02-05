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
	quinePath, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	path, err = ioutil.TempDir("", "quine")
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

var expectedMain = `package main

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

log.SetPrefix(app + ": ")
}

func main() {
// Process flags
parseFlags()
os.Exit(testMain())
}
`

func TestWriteMain(t *testing.T) {
	var buf bytes.Buffer
	tests := []struct {
		license  License
		expected string
	}{
		{None, ""},
		{Apache20, `Copyright [yyyy] [name of copyright owner]

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the spec

`},
		{BSD3Clause, ""},
		{GPL20, `<One line to give the program's name and a brief idea of what it does.>
Copyright (C) <year> <name of author>

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307 USA


`},
		{GPL30, `<one line to give the program's name and a brief idea of what it does.>
 Copyright (C) <year>  <name of author>

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU General Public License as published by
 the Free Software Foundation, either version 3 of the License, or
 (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU General Public License for more details.

 You should have received a copy of the GNU General Public License
 along with this program.  If not, see <http://www.gnu.org/licenses/>.


`},
		{LGPL21, `<one line to give the library's name and an idea of what it does.> Copyright (C) <year> <name of author>

This library is free software; you can redistribute it and/or modify it under the terms of the GNU Lesser General Public License as published by the Free Software Foundation; either version 2.1 of the License, or (at your option) any later version.

This library is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License along with this library; if not, write to the Free Software Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307 USA
`},
		{LGPL30, ""},
		{MIT, ""},
		{MPL20, `This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0. If a copy of the MPL was not distributed with this file, You can obtain one at http://mozilla.org/MPL/2.0/.
`},
	}
	for _, test := range tests {
		licenseType = test.license
		err := writeMain(&buf)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		// check the written file
		b, err := ioutil.ReadFile(filepath.Join(path, mainFile))
		if string(b) != test.expected+expectedMain {
			t.Errorf("got %s\nwant %s", string(b), test.expected+expectedMain)
		}
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

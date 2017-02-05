package main

import "testing"

func TestLicenseFromString(t *testing.T) {
	tests := []struct {
		license  string
		expected License
		err      string
	}{
		{"Apache License 2.0", Apache20, ""},
		{"Apache license 2.0", Apache20, ""},
		{"Apache License 2", Apache20, ""},
		{"Apache-2.0", Apache20, ""},
		{"Apache-2", Apache20, ""},
		{"Apache20", Apache20, ""},
		{"Apache2", Apache20, ""},
		{"apache2", Apache20, ""},
		{"APACHE2", Apache20, ""},
		{"BSD-2-Clause License", BSD2Clause, ""},
		{"BSD-2-clause license", BSD2Clause, ""},
		{"BSD-2-Clause Simplified License", BSD2Clause, ""},
		{"BSD-2-Clause", BSD2Clause, ""},
		{"BSD2Clause", BSD2Clause, ""},
		{"BSD-2", BSD2Clause, ""},
		{"BSD2", BSD2Clause, ""},
		{"bsd2", BSD2Clause, ""},
		{"BSD-3-Clause License", BSD3Clause, ""},
		{"BSD-3-clause license", BSD3Clause, ""},
		{"BSD-3-Clause New or Revised License", BSD3Clause, ""},
		{"BSD-3-Clause", BSD3Clause, ""},
		{"BSD3Clause", BSD3Clause, ""},
		{"BSD-3", BSD3Clause, ""},
		{"BSD3", BSD3Clause, ""},
		{"bsd3", BSD3Clause, ""},
		{"GNU GENERAL PUBLIC LICENSE V2.0 ONLY", GPL20, ""},
		{"Gnu General Public License V2 Only", GPL20, ""},
		{"Gnu General Public License V2.0", GPL20, ""},
		{"Gnu General Public License V2", GPL20, ""},
		{"General Public License V2.0", GPL20, ""},
		{"General Public License V2", GPL20, ""},
		{"GPL-2.0", GPL20, ""},
		{"GPL-20", GPL20, ""},
		{"GPL-2", GPL20, ""},
		{"GPL20", GPL20, ""},
		{"GPL2", GPL20, ""},
		{"GNU GENERAL PUBLIC LICENSE V3.0 ONLY", GPL30, ""},
		{"Gnu General Public License V3 Only", GPL30, ""},
		{"Gnu General Public License V3.0", GPL30, ""},
		{"Gnu General Public License V3", GPL30, ""},
		{"General Public License V3.0", GPL30, ""},
		{"General Public License V3", GPL30, ""},
		{"GPL-3.0", GPL30, ""},
		{"GPL-30", GPL30, ""},
		{"GPL-3", GPL30, ""},
		{"GPL30", GPL30, ""},
		{"GPL3", GPL30, ""},
		{"GNU LESSER GENERAL PUBLIC LICENSE V2.0 ONLY", LGPL20, ""},
		{"gnu lesser general public license v2.0 only", LGPL20, ""},
		{"GNU Lesser General Public License V2.0", LGPL20, ""},
		{"GNU Lesser General Public License V2 Only", LGPL20, ""},
		{"GNU Lesser General Public License V2", LGPL20, ""},
		{"Lesser General Public License V2.0 Only", LGPL20, ""},
		{"Lesser General Public License V2.0", LGPL20, ""},
		{"Lesser General Public License V2 Only", LGPL20, ""},
		{"Lesser General Public License V2", LGPL20, ""},
		{"LGPL-2.0", LGPL20, ""},
		{"LGPL-2", LGPL20, ""},
		{"LGPL20", LGPL20, ""},
		{"LGPL2", LGPL20, ""},
		{"GNU LESSER GENERAL PUBLIC LICENSE V2.1 ONLY", LGPL21, ""},
		{"gnu lesser general public license v2.1 only", LGPL21, ""},
		{"GNU Lesser General Public License V2.1", LGPL21, ""},
		{"Lesser General Public License V2.1 Only", LGPL21, ""},
		{"Lesser General Public License V2.1", LGPL21, ""},
		{"LGPL-2.1", LGPL21, ""},
		{"LGPL21", LGPL21, ""},
		{"GNU LESSER GENERAL PUBLIC LICENSE V3.0 ONLY", LGPL30, ""},
		{"gnu lesser general public license v3.0 only", LGPL30, ""},
		{"GNU Lesser General Public License V3.0", LGPL30, ""},
		{"GNU Lesser General Public License V3 Only", LGPL30, ""},
		{"GNU Lesser General Public License V3", LGPL30, ""},
		{"Lesser General Public License V3.0 Only", LGPL30, ""},
		{"Lesser General Public License V3.0", LGPL30, ""},
		{"Lesser General Public License V3 Only", LGPL30, ""},
		{"Lesser General Public License V3", LGPL30, ""},
		{"LGPL-3.0", LGPL30, ""},
		{"LGPL-3", LGPL30, ""},
		{"LGPL30", LGPL30, ""},
		{"LGPL3", LGPL30, ""},
		{"MIT License", MIT, ""},
		{"MIT", MIT, ""},
		{"Mozilla Public License 2.0", MPL20, ""},
		{"Mozilla Public License 2", MPL20, ""},
		{"MPL-2.0", MPL20, ""},
		{"MPL-2", MPL20, ""},
		{"MPL20", MPL20, ""},
		{"MPL2", MPL20, ""},
		{"", None, ""},
		{"fdas", None, "unsupported license: fdas"},
		{"Gen Public License", None, "unsupported license: Gen Public License"},
	}

	for _, test := range tests {
		l, err := LicenseFromString(test.license)
		if err != nil {
			if err.Error() != test.err {
				t.Errorf("got %s; want %s", err.Error(), test.err)
			}
			continue
		}
		if test.err != "" {
			t.Errorf("got no error; want %s", test.err)
			continue
		}
		if l != test.expected {
			t.Errorf("got %s; want %s", l, test.expected)
		}
	}
}

func TestID(t *testing.T) {
	tests := []struct {
		l        License
		expected string
	}{
		{None, "None"},
		{Apache20, "Apache-2.0"},
		{BSD2Clause, "BSD-2-Clause"},
		{BSD3Clause, "BSD-3-Clause"},
		{GPL20, "GPL-2.0"},
		{GPL30, "GPL-3.0"},
		{LGPL20, "LGPL-2.0"},
		{LGPL21, "LGPL-2.1"},
		{LGPL30, "LGPL-3.0"},
		{MIT, "MIT"},
		{MPL20, "MPL-2.0"},
	}
	for _, test := range tests {
		s := test.l.String()
		if s != test.expected {
			t.Errorf("got %q; want %q", s, test.expected)
		}
	}
}

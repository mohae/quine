package main

import (
	"strconv"
	"strings"
)

// Constants for OSI Approved License using their SPDX short code.
const (
	None License = iota
	Apache20
	BSD2Clause
	BSD3Clause
	GPL20
	GPL30
	LGPL20
	LGPL21
	LGPL30
	MIT
	MPL20
)

// License represents an OSI Approved License.
type License int

// Returns the SPDX Identifier (short code).
func (l License) String() string {
	return l.ID()
}

// Returns the SPDX Identifier (short code).
func (l License) ID() string {
	switch l {
	case Apache20:
		return "Apache-2.0"
	case BSD2Clause:
		return "BSD-2-Clause"
	case BSD3Clause:
		return "BSD-3-Clause"
	case GPL20:
		return "GPL-2.0"
	case GPL30:
		return "GPL-3.0"
	case LGPL20:
		return "LGPL-2.0"
	case LGPL21:
		return "LGPL-2.1"
	case LGPL30:
		return "LGPL-3.0"
	case MIT:
		return "MIT"
	case MPL20:
		return "MPL-2.0"
	case None:
		return "None"
	default:
		return strconv.Itoa(int(l))
	}
}

// UnsupportedLicenseErr occurs when a string cannot be matched with a quine
// supported license.
type UnsupportedLicenseErr struct {
	s string
}

func (e UnsupportedLicenseErr) Error() string {
	return "unsupported license: " + e.s
}

// LicenseFromString will return the license associated with a string. The
// string may be the license's full name or identifier (short code) as listed
// by SPDX,org. Variants of the short code that drop the punctuation and/or any
// trailing 0 are also accepted. For full names, the license name should not
// include any quotes: e,g, BSD 2-clause Simplified License not BSD 2-clause
// "Simplified" License.
//
// An empty string will result in `None` and is not considered an error. Any
// license text that can't be matched will result in an error.
//
// All text is upper-cased before comparisons.
func LicenseFromString(s string) (License, error) {
	if s == "" { // nothing specified is not an error state.
		return None, nil
	}

	v := strings.ToUpper(s)
	switch v {
	case "APACHE LICENSE 2.0", "APACHE LICENSE 2", "APACHE-2.0", "APACHE-2", "APACHE20", "APACHE2":
		return Apache20, nil
	case "BSD-2-CLAUSE LICENSE", "BSD-2-CLAUSE SIMPLIFIED LICENSE", "BSD-2-CLAUSE", "BSD2CLAUSE", "BSD-2", "BSD2":
		return BSD2Clause, nil
	case "BSD-3-CLAUSE LICENSE", "BSD-3-CLAUSE NEW OR REVISED LICENSE", "BSD-3-CLAUSE", "BSD3CLAUSE", "BSD-3", "BSD3":
		return BSD3Clause, nil
	case "GNU GENERAL PUBLIC LICENSE V2.0 ONLY", "GNU GENERAL PUBLIC LICENSE V2 ONLY", "GNU GENERAL PUBLIC LICENSE V2.0", "GNU GENERAL PUBLIC LICENSE V2",
		"GENERAL PUBLIC LICENSE V2.0 ONLY", "GENERAL PUBLIC LICENSE V2.0", "GENERAL PUBLIC LICENSE V2 ONLY", "GENERAL PUBLIC LICENSE V2",
		"GPL-2.0", "GPL-2", "GPL-20", "GPL20", "GPL2":
		return GPL20, nil
	case "GNU GENERAL PUBLIC LICENSE V3.0 ONLY", "GNU GENERAL PUBLIC LICENSE V3 ONLY", "GNU GENERAL PUBLIC LICENSE V3.0", "GNU GENERAL PUBLIC LICENSE V3",
		"GENERAL PUBLIC LICENSE V3.0 ONLY", "GENERAL PUBLIC LICENSE V3.0", "GENERAL PUBLIC LICENSE V3 ONLY", "GENERAL PUBLIC LICENSE V3",
		"GPL-3.0", "GPL-3", "GPL-30", "GPL30", "GPL3":
		return GPL30, nil
	case "GNU LESSER GENERAL PUBLIC LICENSE V2.0 ONLY", "GNU LESSER GENERAL PUBLIC LICENSE V2 ONLY", "GNU LESSER GENERAL PUBLIC LICENSE V2.0", "GNU LESSER GENERAL PUBLIC LICENSE V2",
		"LESSER GENERAL PUBLIC LICENSE V2.0 ONLY", "LESSER GENERAL PUBLIC LICENSE V2 ONLY", "LESSER GENERAL PUBLIC LICENSE V2.0", "LESSER GENERAL PUBLIC LICENSE V2",
		"LGPL-2.0", "LGPL-2", "LGPL20", "LGPL2":
		return LGPL20, nil
	case "GNU LESSER GENERAL PUBLIC LICENSE V2.1 ONLY", "GNU LESSER GENERAL PUBLIC LICENSE V2.1", "LESSER GENERAL PUBLIC LICENSE V2.1 ONLY", "LESSER GENERAL PUBLIC LICENSE V2.1",
		"LGPL-2.1", "LGPL21":
		return LGPL21, nil
	case "GNU LESSER GENERAL PUBLIC LICENSE V3.0 ONLY", "GNU LESSER GENERAL PUBLIC LICENSE V3 ONLY", "GNU LESSER GENERAL PUBLIC LICENSE V3.0", "GNU LESSER GENERAL PUBLIC LICENSE V3",
		"LESSER GENERAL PUBLIC LICENSE V3.0 ONLY", "LESSER GENERAL PUBLIC LICENSE V3 ONLY", "LESSER GENERAL PUBLIC LICENSE V3.0", "LESSER GENERAL PUBLIC LICENSE V3",
		"LGPL-3.0", "LGPL-3", "LGPL30", "LGPL3":
		return LGPL30, nil
	case "MIT LICENSE", "MIT":
		return MIT, nil
	case "MOZILLA PUBLIC LICENSE 2.0", "MOZILLA PUBLIC LICENSE 2", "MPL-2.0", "MPL-2", "MPL20", "MPL2":
		return MPL20, nil
	default:
		return None, UnsupportedLicenseErr{s}
	}
}

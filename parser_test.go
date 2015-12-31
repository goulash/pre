// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package pre

import (
	"path/filepath"
	"testing"

	"github.com/goulash/osutil"
)

const (
	testExt   = "test"
	resultExt = "result"
)

// proc is the default processor for our tests.
var proc *Processor

func TestParser(z *testing.T) {
	proc = New()
	proc.AddCommenter(CComment, true)
	proc.AddCommenter(CppComment, true)

	matches, err := filepath.Glob("testdata/*." + testExt)
	if err != nil {
		z.Fatal(err)
	}
	for _, m := range matches {
		testParser(z, m)
	}
}

func testParser(z *testing.T, path string) {
	result := path[:len(path)-len(testExt)] + resultExt
	ex, err := osutil.FileExists(result)
	if err != nil {
		z.Fatal(err)
	}
	if !ex {
		z.Fatalf("missing result file: %s", result)
	}

	n, err := proc.Parse(path)
	if err != nil {
		z.Error(err)
		return
	}
	r, err := proc.Parse(result)
	if err != nil {
		z.Error(err)
		return
	}

	ns, rs := n.String(), r.String()
	if ns == "" || rs == "" {
		z.Errorf("neither test nor exp. result should be empty string")
	} else if ns != rs {
		z.Errorf("processed test did not match result\nGOT:\n%s\n\nEXPECTED:\n%s\n", ns, rs)
	}
}

var tests = []struct {
	Test string
	Exp  string
}{
	{"// Comments will be stripped\nBut the rest of the file should remain.\n",
		"\nBut the rest of the file should remain.\n"},
}

func TestSimple(z *testing.T) {
	p := New()
	p.AddCommenter(CppComment, true)

	for _, t := range tests {
		n, err := p.ParseString("internal", t.Test)
		if err != nil {
			z.Error(err)
		}
		if n.Len() == 0 {
			z.Errorf("ParseString(%q).Len() == 0, want > 0", t.Test)
			continue
		}
		if n.String() != t.Exp {
			z.Errorf("ParseString(%q) = %q, want %q", t.Test, n.String(), t.Exp)
		}
	}
}

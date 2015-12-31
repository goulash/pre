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

	if n.String() != r.String() {
		z.Errorf("processed test did not match result\nGOT:\n%s\n\nEXPECTED:\n%s\n", n.String(), r.String())
	}
}

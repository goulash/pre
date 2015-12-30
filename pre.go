// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Package pre offers preprocessing of files.
//
// Commands available are:
//
//  printf
//  include
//  require
//  define
//  ifdef
//  ifndef
package pre

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Processor struct {
	Rune rune
}

func New() *Processor {
	return &Processor{
		Rune: '#',
	}
}

func (p *Processor) Parse(path string) (*Text, error) {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintln("Warning:", err)
		abs = path
	}
	fp, err = filepath.EvalSymlinks(abs)
	if err != nil {
		fmt.Fprintln("Warning:", err)
		fp = path
	}
	return ParseString(path, fp, string(bs))
}

func (p *Processor) ParseString(name, path, data string) (*Text, error) {
	f := &File{name: name, path: path, data: data, base: 0, fpos: 0}
	text := &Text{
		fset:  &FileSet{[]*File{f}},
		files: map[string]*File{path: f},
	}

	p := parser{p.Rune, text}
	err := p.Parse(f, 0)
	if err != nil {
		return nil, err
	}
	return p.text, nil
}

type Text struct {
	buf   bytes.Buffer     // contains the processed text
	fset  *FileSet         // describes the data in buf, what belongs to what file
	files map[string]*File // a map of all loaded files: path -> *File
}

func (t Text) String() string           { return t.buf.String() }
func (t Text) Bytes() []byte            { return t.buf.Bytes() }
func (t Text) PosInfo(pos Pos) *PosInfo { return t.fset.PosInfo(pos) }

type Pos int

type File struct {
	name string // name of file, as refered to from outside
	path string // absolute path to file, as far as possible
	data string // file data
	base Pos    // position in global processed text
	fpos Pos    // position in file
}

// The FileSet object is used solely to calculate original positions,
// so we can get position information for a certain offset.
type FileSet struct {
	files []*File
}

type PosInfo struct {
	Name   string
	Line   int
	Column int
}

// PosInfo maps the position from the resulting text to the original file.
func (f FileSet) PosInfo(pos Pos) *PosInfo {
	var prev *File
	for _, f := range f.files {
		if pos < f.base {
			break
		}
		prev = f
	}

	offset := f.fpos + (pos - f.base) + 1 // +1 because we are interested in that point
	data := f.data[:offset]
	pinfo := &PosInfo{
		Name: f.name,
		Line: 1 + strings.Count(data, "\n"),
	}
	if i := strings.LastIndex(data, "\n"); i >= 0 {
		pinfo.Column = offset - i
	} else {
		pinfo.Column = 1 + len(data)
	}
	return pinfo
}

type parser struct {
	trig rune
	text *Text

	code string
	i    int
}

func (p *parser) Parse(f *File, offset int) error {
}

func (p *parser) next() (Node, error) {
	if p.i == len(p.code) {
		return nil, io.EOF
	}

	// TODO: finish this

	r, size := utf8.DecodeRuneInString(p.code[p.i:])
	for r == ';' || unicode.IsSpace(r) {
		p.i += size
		if r == ';' {
			for p.i < len(p.code) && r != '\n' {
				r, size = utf8.DecodeRuneInString(p.code[p.i:])
				p.i += size
			}
		}
		if p.i == len(p.code) {
			return nil, io.EOF
		}
		r, size = utf8.DecodeRuneInString(p.code[p.i:])
	}
}

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

import "github.com/goulash/pre/ast"

type Processor struct {
	// Trigger is the string which begins an action (command).
	// The default trigger is "#", which is the same as the C/C++ pre-processor.
	Trigger string

	// MaxIncludeDepth is the maximum number of nested includes that can occur
	// before an error is thrown. This is to prevent infinite include loops.
	MaxIncludeDepth int

	// Commenters define what kind of comments are accepted in the parsed text.
	// Triggers are ignored when they are inside a comment. Comments can also
	// be stripped out of the text, or just left there.
	Commenters ast.Commenters
}

func New() *Processor {
	return &Processor{
		Trigger:         "#",
		MaxIncludeDepth: 128,
	}
}

func (p *Processor) AddCommenter(c *ast.Commenter, strip bool) {
	c.Strip = strip
	p.Commenters = append(p.Commenters, c)
}

func (p *Processor) Parse(path string) (ast.Node, error) {
	parser := newParser(p)
	err := parser.Parse(path)
	nod := parser.Root()
	return nod, err
}

func (p *Processor) ParseString(name, code string) (ast.Node, error) {
	parser := newParser(p)
	err := parser.ParseString(name, code)
	nod := parser.Root()
	return nod, err
}

func newParser(p *Processor) *ast.Parser {
	return &ast.Parser{
		Trigger:         p.Trigger,
		MaxIncludeDepth: p.MaxIncludeDepth,
		Commenters:      p.Commenters,
	}
}

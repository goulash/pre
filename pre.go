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

func (p *Processor) Parse(path string) (ast.Node, error) {
	return nil, nil
}

func (p *Processor) ParseString(name, code string) (ast.Node, error) {
	return nil, nil
}

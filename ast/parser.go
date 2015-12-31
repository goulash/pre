// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package ast

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/goulash/lex"
)

var (
	ErrMaxDepthExceeded = errors.New("maximum include depth exceeded")

	errRequireIgnore = errors.New("ignoring file because already read")
)

type parseFn func(*lex.Reader) (parseFn, error)

type Parser struct {
	// trigger and commenters are used by the lex* methods
	Trigger    string
	Commenters Commenters

	MaxIncludeDepth int

	nod          *FileNode
	files        map[string]bool // included file paths
	includeDepth int             // include depth
}

func (p *Parser) Parse(path string) error {
	return p.parseFile(path, PosInfo{Name: path}, true)
}

func (p *Parser) ParseString(name, code string) (err error) {
	p.nod = &FileNode{
		PosInfo: PosInfo{Name: name},
		name:    name,
		path:    "",
		root:    nil,
	}
	r := lex.NewReader(lex.Lex(name, string(code), p.lexText))
	for fn := p.parseNext; fn != nil; {
		fn, err = fn(r)
		if err != nil {
			break
		}
	}
	return
}

func (p *Parser) Root() *FileNode {
	return p.nod
}

func (p *Parser) parseFile(name string, pi PosInfo, unique bool) (err error) {
	if p.includeDepth >= p.MaxIncludeDepth {
		return ErrMaxDepthExceeded
	}

	bs, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(name)
	if err != nil {
		// TODO: should I do this?
		fmt.Fprintln(os.Stderr, "Warning:", err)
		abs = name
	}
	path, err := filepath.EvalSymlinks(abs)
	if err != nil {
		// TODO: should I do this?
		fmt.Fprintln(os.Stderr, "Warning:", err)
		path = abs
	}

	// Note: this is currently best-effort. If same files are
	// mounted in different places, we will not catch it. But
	// then again, maybe we should just accept that.
	if unique {
		if p.files[path] {
			// We already read this file, ignore it.
			return errRequireIgnore
		}
		p.files[path] = true
	}

	fn := &FileNode{
		PosInfo: pi,
		name:    name,
		path:    path,
		root:    p.nod,
	}
	if p.nod != nil {
		p.nod.addNode(fn)
	}
	p.nod = fn
	p.includeDepth++
	r := lex.NewReader(lex.Lex(name, string(bs), p.lexText))
	for fn := p.parseNext; fn != nil; {
		fn, err = fn(r)
		if err != nil && err != errRequireIgnore {
			break
		}
	}
	p.includeDepth--
	if p.nod.root != nil {
		p.nod = p.nod.root
	}
	return
}

func (p *Parser) parseNext(r *lex.Reader) (parseFn, error) {
	tok := r.Peek()
	switch tok.Type {
	case TypeText:
		return p.parseText, nil
	case TypeComment:
		return p.parseComment, nil
	case TypeActionBegin:
		return p.parseAction, nil
	case lex.TypeError:
		return nil, errors.New(tok.Value)
	case lex.TypeEOF:
		return nil, nil
	default:
		// TODO: what kind of token was unexpected?
		return nil, errors.New("unexpected token")
	}
}

func (p *Parser) parseText(r *lex.Reader) (parseFn, error) {
	t := r.Next()
	p.nod.addNode(&TextNode{posInfo(r), t.Value})
	return p.parseNext, nil
}

func (p *Parser) parseComment(r *lex.Reader) (parseFn, error) {
	t := r.Next()
	p.nod.addNode(&CommentNode{posInfo(r), t.Value, p.Commenters.First(t.Value)})
	return p.parseNext, nil
}

func (p *Parser) parseAction(r *lex.Reader) (parseFn, error) {
	tok := r.Next()
	if tok.Type != TypeIdent {
		return nil, errors.New("expecting command identifier")
	}

	switch cmd := tok.Value; cmd {
	case "include":
		return p.parseCmdInclude, nil
	case "require":
		return p.parseCmdRequire, nil
	case "error":
		return p.parseCmdError, nil
	default:
		return nil, fmt.Errorf("unknown command %s", cmd)
	}
}

func (p *Parser) parseCmdInclude(r *lex.Reader) (parseFn, error) {
	pi := posInfo(r)
	args, ok := r.Expect(TypeString, TypeActionEnd)
	if !ok {
		return nil, fmt.Errorf("command include takes a single string argument")
	}

	return p.parseNext, p.parseFile(args[0].Value, pi, false)
}

// this is best effort require at the moment. There are several ways to work around this.
func (p *Parser) parseCmdRequire(r *lex.Reader) (parseFn, error) {
	pi := posInfo(r)
	args, ok := r.Expect(TypeString, TypeActionEnd)
	if !ok {
		return nil, fmt.Errorf("command require takes a single string argument")
	}

	return p.parseNext, p.parseFile(args[0].Value, pi, true)
}

func (p *Parser) parseCmdError(r *lex.Reader) (parseFn, error) {
	pi := posInfo(r)
	args, ok := r.Expect(TypeString, TypeActionEnd)
	if !ok {
		return nil, fmt.Errorf("command error takes a single string argument")
	}

	return nil, fmt.Errorf("%s: %s", pi, args[0].Value)
}

func posInfo(r *lex.Reader) PosInfo {
	n, l, c := r.PosInfo()
	return PosInfo{n, l, c}
}

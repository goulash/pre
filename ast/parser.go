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

type Error struct {
	Err     error
	PosInfo PosInfo
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %v", e.PosInfo, e.Err)
}

type Parser struct {
	Trigger         string
	Commenters      Commenters
	MaxIncludeDepth int

	nod          *FileNode
	files        map[string]bool // included file paths
	includeDepth int             // include depth
}

// Root returns the root node in the AST.
func (p *Parser) Root() *FileNode {
	return p.nod
}

// Parse parses a file and returns an error if one occurs.
func (p *Parser) Parse(path string) error {
	return p.parseFile(path, PosInfo{Name: path}, true)
}

// ParseString parses a string as the root node.
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
		if err != nil && err != errRequireIgnore {
			break
		}
	}
	if err != nil {
		err = &Error{err, posInfo(r)}
	}
	return
}

type parseFn func(*lex.Reader) (parseFn, error)

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
		if p.files == nil {
			p.files = make(map[string]bool)
		} else if p.files[path] {
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
	if err != nil {
		err = &Error{err, posInfo(r)}
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
	case typeText:
		return p.parseText, nil
	case typeComment:
		return p.parseComment, nil
	case typeActionBegin:
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

func (p *Parser) parseShebang(r *lex.Reader) (parseFn, error) {
	_, ok := r.Expect(typeExclamation, typeSlash)
	pi := posInfo(r)
	if !ok {
		return nil, errors.New("shebang paths are absolute, expecting slash '/'")
	}

	if pi.Line != 1 {
		return nil, errors.New("shebang only valid on first line of file")
	}

	for tok := r.Next(); tok.Type != typeActionEnd; tok = r.Next() {
		// shebang has nothing to do with us, so we consume until it's over.
		if tok.Type == lex.TypeEOF {
			return nil, errors.New("unexpected EOF")
		}
	}
	return p.parseNext, nil
}

func (p *Parser) parseAction(r *lex.Reader) (parseFn, error) {
	r.Next() // trigger token

	// If the token afterwards is !, then it could be something like #!/usr/bin/env
	if r.Peek().Type == typeExclamation {
		return p.parseShebang, nil
	}

	tok := r.Next()
	if tok.Type != typeIdent {
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
	args, ok := r.Expect(typeString, typeActionEnd)
	if !ok {
		return nil, errors.New("command include takes a single string argument")
	}

	path := filepath.Join(filepath.Dir(p.nod.name), args[0].Value)
	return p.parseNext, p.parseFile(path, pi, false)
}

// this is best effort require at the moment. There are several ways to work around this.
func (p *Parser) parseCmdRequire(r *lex.Reader) (parseFn, error) {
	pi := posInfo(r)
	args, ok := r.Expect(typeString, typeActionEnd)
	if !ok {
		return nil, errors.New("command require takes a single string argument")
	}

	path := filepath.Join(filepath.Dir(p.nod.name), args[0].Value)
	return p.parseNext, p.parseFile(path, pi, true)
}

func (p *Parser) parseCmdError(r *lex.Reader) (parseFn, error) {
	args, ok := r.Expect(typeString, typeActionEnd)
	if !ok {
		return nil, errors.New("command error takes a single string argument")
	}

	return nil, errors.New(args[0].Value)
}

func posInfo(r *lex.Reader) PosInfo {
	n, l, c := r.PosInfo()
	return PosInfo{n, l, c}
}

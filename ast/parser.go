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

// We limit the recursion depth of includes to an arbitrary value
// to catch infinite-loops.
var MaxIncludeDepth int = 128

var (
	ErrRequireIgnore = errors.New("ignoring file because already read")
)

type parseFn func(*lex.Reader) (parseFn, error)

type parser struct {
	// trigger and commenters are used by the lex* methods
	trigger    string
	commenters Commenters

	nod *FileNode

	files map[string]bool // included file paths
	depth int             // include depth
}

func (p *parser) parseFile(name string, pi PosInfo, unique bool) (err error) {
	if p.depth >= MaxIncludeDepth {
		return errors.New("maximum include depth reached")
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
			return ErrRequireIgnore
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
	p.depth++
	r := lex.NewReader(lex.Lex(name, string(bs), p.lexText))
	for fn := p.parseNext; fn != nil; {
		fn, err = fn(r)
		if err != nil {
			break
		}
	}
	p.depth--
	if p.nod.root != nil {
		p.nod = p.nod.root
	}
	return
}

func (p *parser) parseNext(r *lex.Reader) (parseFn, error) {
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

func (p *parser) parseText(r *lex.Reader) (parseFn, error) {
	t := r.Next()
	p.nod.addNode(&TextNode{posInfo(r), t.Value})
	return p.parseNext, nil
}

func (p *parser) parseComment(r *lex.Reader) (parseFn, error) {
	t := r.Next()
	p.nod.addNode(&CommentNode{posInfo(r), t.Value, p.commenters.First(t.Value)})
	return p.parseNext, nil
}

func (p *parser) parseAction(r *lex.Reader) (parseFn, error) {
	tok := r.Next()
	if tok.Type != TypeIdent {
		return nil, errors.New("expecting command identifier")
	}

	switch cmd := tok.Value; cmd {
	case "include":
		return p.parseInclude, nil
	case "require":
		return p.parseRequire, nil
	default:
		return nil, fmt.Errorf("unknown command %s", cmd)
	}
}

func (p *parser) parseInclude(r *lex.Reader) (parseFn, error) {
	pi := posInfo(r)
	arg := r.Next()
	end := r.Next()
	if arg.Type != TypeString || end.Type != TypeActionEnd {
		return nil, fmt.Errorf("command include takes a single string argument")
	}

	return p.parseNext, p.parseFile(arg.Value, pi, false)
}

// this is best effort require at the moment. There are several ways to work around this.
func (p *parser) parseRequire(r *lex.Reader) (parseFn, error) {
	pi := posInfo(r)
	arg := r.Next()
	end := r.Next()
	if arg.Type != TypeString || end.Type != TypeActionEnd {
		return nil, fmt.Errorf("command require takes a single string argument")
	}

	return p.parseNext, p.parseFile(arg.Value, pi, true)
}

func posInfo(r *lex.Reader) PosInfo {
	n, l, c := r.PosInfo()
	return PosInfo{n, l, c}
}

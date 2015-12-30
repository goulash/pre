// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package ast

import (
	"bytes"
	"fmt"
	"strings"
)

type Node interface {
	Type() NodeType
	String() string
	Pos() *PosInfo
	Len() int
	Offset(offset int) *PosInfo
}

type NodeType int

const (
	ErrorType NodeType = iota
	FileType
	TextType
	CommentType
)

// PosInfo {{{

type PosInfo struct {
	Name   string
	Line   int
	Column int
}

func (p PosInfo) Pos() *PosInfo { return &p }

func (p PosInfo) String() string {
	return fmt.Sprintf("%s:%d:%d", p.Name, p.Line, p.Column)
}

func (p PosInfo) OffsetIn(data string, offset int) *PosInfo {
	if offset > len(data) {
		return nil
	}
	code := data[:offset]
	pi := &PosInfo{
		Name: p.Name,
		Line: p.Line + strings.Count(code, "\n"),
	}
	if i := strings.LastIndex(code, "\n"); i >= 0 {
		pi.Column = offset - i
	} else {
		pi.Column = 1 + len(code)
	}
	return pi
}

// }}}

// TextNode {{{

type TextNode struct {
	PosInfo
	val string
}

func (n TextNode) Type() NodeType             { return TextType }
func (n TextNode) String() string             { return n.val }
func (n TextNode) Len() int                   { return len(n.val) }
func (n TextNode) Offset(offset int) *PosInfo { return n.OffsetIn(n.val, offset) }

// }}}

// CommentNode {{{

type CommentNode struct {
	PosInfo
	val string
	c   *Commenter
}

func (n CommentNode) Type() NodeType             { return CommentType }
func (n CommentNode) String() string             { return n.val }
func (n CommentNode) Len() int                   { return len(n.val) }
func (n CommentNode) Offset(offset int) *PosInfo { return n.OffsetIn(n.val, offset) }

// }}}

// FileNode {{{

type FileNode struct {
	PosInfo
	name  string
	path  string
	root  *FileNode
	nodes []Node
}

func (n FileNode) Type() NodeType { return FileType }

func (fn FileNode) String() string {
	var buf bytes.Buffer
	for _, n := range fn.nodes {
		buf.WriteString(n.String())
	}
	return buf.String()
}

func (fn FileNode) Len() int {
	var total int
	for _, n := range fn.nodes {
		total += n.Len()
	}
	return total
}

func (fn FileNode) Offset(offset int) *PosInfo {
	for _, n := range fn.nodes {
		pi := n.Offset(offset)
		if pi != nil {
			return pi
		}
		offset -= n.Len()
	}
	return nil
}

func (fn FileNode) Nodes() []Node {
	var nodes []Node
	for _, n := range fn.nodes {
		if n.Type() == FileType {
			nodes = append(nodes, n.(*FileNode).Nodes()...)
			continue
		}
		nodes = append(nodes, n)
	}
	return nodes
}

func (fn *FileNode) addNode(n Node) {
	fn.nodes = append(fn.nodes, n)
}

// }}}

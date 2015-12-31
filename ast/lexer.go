// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package ast

import "github.com/goulash/lex"

const (
	// We continue where the reserved types left off
	typeText lex.Type = (lex.TypeEOF + 1) + iota
	typeComment

	typeExclamation // '!'
	typeSlash       // '/'

	typeActionBegin
	typeActionEnd
	typeIdent
	typeString
)

// lexText scans until an action of the end of the text.
// lexText expects to start at the beginning of a line.
func (p *Parser) lexText(l *lex.Lexer) lex.StateFn {
	for {
		n := l.AcceptRun(lex.Space)
		// We accept the trigger if the rune before the whitespace is a newline.
		if l.HasPrefix(p.Trigger) && (l.Pos() == n || l.Input(-n - 1)[0] == '\n') {
			l.Dec(n) // don't include leading space in text
			if l.Len() > 0 {
				l.Emit(typeText)
			}
			l.Inc(n)
			l.Ignore()
			return p.lexActionBegin
		}
		if p.Commenters.IsComment(l.Input(0)) {
			if l.Len() > 0 {
				l.Emit(typeText)
			}
			return p.lexComment
		}

		// We don't have a space, it's not a comment or trigger, so make sure
		// it's not an EOF. Otherwise, we will move on to the next rune.
		if l.Next() == lex.EOF {
			break
		}
	}
	// Correctly reached EOF.
	if l.Len() > 0 {
		l.Emit(typeText)
	}
	l.Emit(lex.TypeEOF)
	return nil
}

// lexComment scans a comment, because the trigger doesn't count in a comment.
// The comment includes the //, /* */, or whatever.
func (p *Parser) lexComment(l *lex.Lexer) lex.StateFn {
	// Find out which kind of comment we have, so we know how to deal with it.
	c := p.Commenters.First(l.Input(0))

	l.Inc(len(c.Begin))
	var end = c.End
	if end == "" {
		end = "\n"
	}
	for !l.Consume(end) && l.Next() != lex.EOF {
		// absorb as long as we don't hit EOF or end-of-comment
	}
	if c.End == "" {
		l.Dec(1)
	}

	if c.Strip {
		l.Ignore()
	} else {
		l.Emit(typeComment)
	}
	// If we exited because of EOF, then Peek will also return EOF.
	if l.Peek() == lex.EOF {
		l.Emit(lex.TypeEOF)
		return nil
	}
	return p.lexText
}

func (p *Parser) lexActionBegin(l *lex.Lexer) lex.StateFn {
	l.Inc(len(p.Trigger))
	l.Emit(typeActionBegin)
	return p.lexInsideAction
}

func (p *Parser) lexActionEnd(l *lex.Lexer) lex.StateFn {
	if !(l.Consume("\n") || l.Consume("\r\n")) {
		return l.Errorf("malformed end-of-line")
	}
	l.Emit(typeActionEnd)
	return p.lexText
}

// lexSpace scans all spaces. One space may have already been read.
// It does not emit any space tokens however. We don't have a use for that yet.
func (p *Parser) lexSpace(l *lex.Lexer) lex.StateFn {
	l.AcceptFuncRun(lex.IsSpace)
	l.Ignore()
	return p.lexInsideAction
}

// lexQuote scans all the string inside a quote.
// Only double-quote is supported at the moment.
func (p *Parser) lexQuote(l *lex.Lexer) lex.StateFn {
	// lexQuote is called for ', ", and `.
	if l.Next() != '"' {
		return l.Errorf("only support double-quoted strings")
	}
	l.Ignore()

loop:
	for {
		switch l.Next() {
		case '\\':
			if r := l.Next(); r != lex.EOF && r != '\n' {
				break
			}
			fallthrough
		case lex.EOF, '\n':
			return l.Errorf("unterminated quoted string")
		case '"':
			break loop
		}
	}
	l.Dec(1)
	l.Emit(typeString)
	l.Inc(1)
	l.Ignore()
	return p.lexInsideAction
}

func (p *Parser) lexInsideAction(l *lex.Lexer) lex.StateFn {
	switch r := l.Peek(); {
	case lex.IsEndline(r):
		return p.lexActionEnd
	case lex.IsSpace(r):
		return p.lexSpace
	case lex.IsQuote(r):
		return p.lexQuote
	case lex.IsAlphaNumeric(r):
		return p.lexAlphaNumeric
	case r == '!':
		l.Next()
		l.Emit(typeExclamation)
		return p.lexInsideAction
	case r == '/':
		l.Next()
		l.Emit(typeSlash)
		return p.lexInsideAction
	case r == lex.EOF:
		return l.Errorf("unexpected EOF")
	default:
		return l.Errorf("unexpected rune: %v", r)
	}
}

func (p *Parser) lexAlphaNumeric(l *lex.Lexer) lex.StateFn {
	l.AcceptFuncRun(lex.IsAlphaNumeric)
	l.Emit(typeIdent)
	return p.lexInsideAction
}

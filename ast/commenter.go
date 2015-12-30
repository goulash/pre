// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package ast

import "strings"

type Commenter struct {
	Begin string
	End   string

	// If Strip is true, the comment is stripped out of the text.
	Strip bool
}

func (c *Commenter) IsComment(s string) bool {
	return strings.HasPrefix(s, c.Begin)
}

type Commenters []*Commenter

func (cs Commenters) IsComment(s string) bool {
	for _, c := range cs {
		if c.IsComment(s) {
			return true
		}
	}
	return false
}

func (cs Commenters) First(s string) *Commenter {
	for _, c := range cs {
		if c.IsComment(s) {
			return c
		}
	}
	return nil
}

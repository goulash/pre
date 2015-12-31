// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package pre

import "github.com/goulash/pre/ast"

var (
	LispComment = PrefixCommenter(";")
	CppComment  = PrefixCommenter("//")
	CComment    = &Commenter{"/*", "*/"}
)

func PrefixCommenter(prefix string) *ast.Commenter {
	return &ast.Commenter{
		Begin: prefix,
	}
}

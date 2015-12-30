// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package pre

var (
	LispComment = PrefixCommenter(";")
	CppComment  = PrefixCommenter("//")
	CComment    = &Commenter{"/*", "*/"}
)

func PrefixCommenter(prefix string) *Commenter {
	return &Commenter{
		Begin: prefix,
		End:   "\n",
	}
}

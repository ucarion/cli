package didyoumean

import (
	"github.com/ucarion/cli/internal/cmdtree"
)

func DidYouMean(tree cmdtree.CommandTree, s string) string {
	var out string
	var best int

	for key := range tree.Children {
		d := distance(s, key)
		if best == 0 || d < best {
			out = key
			best = d
		}
	}

	return out
}

// distance computes a levenstein distance between two strings.
func distance(a, b string) int {
	// The rest of this code is copied from:
	//
	// https://github.com/agnivade/levenshtein/blob/63d27aaa4f5268ad5d467c8a9ab951539d3c6723/levenshtein.go
	//
	// That code is licensed under MIT. To avoid having to take on a dependency,
	// we just copy the code here.

	// We need to convert to []rune if the strings are non-ASCII.
	// This could be avoided by using utf8.RuneCountInString
	// and then doing some juggling with rune indices,
	// but leads to far more bounds checks. It is a reasonable trade-off.
	s1 := []rune(a)
	s2 := []rune(b)

	// swap to save some memory O(min(a,b)) instead of O(a)
	if len(s1) > len(s2) {
		s1, s2 = s2, s1
	}
	lenS1 := len(s1)
	lenS2 := len(s2)

	// init the row
	x := make([]uint16, lenS1+1)
	// we start from 1 because index 0 is already 0.
	for i := 1; i < len(x); i++ {
		x[i] = uint16(i)
	}

	// make a dummy bounds check to prevent the 2 bounds check down below.
	// The one inside the loop is particularly costly.
	_ = x[lenS1]
	// fill in the rest
	for i := 1; i <= lenS2; i++ {
		prev := uint16(i)
		for j := 1; j <= lenS1; j++ {
			current := x[j-1] // match
			if s2[i-1] != s1[j-1] {
				current = min(min(x[j-1]+1, prev+1), x[j]+1)
			}
			x[j-1] = prev
			prev = current
		}
		x[lenS1] = prev
	}
	return int(x[lenS1])
}

func min(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}

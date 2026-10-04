package fribidi

import (
	"fmt"
	"reflect"
	"testing"
	"unicode"

	"golang.org/x/text/unicode/bidi"
)

func TestBracketsTable(t *testing.T) {
	t.Parallel()
	for r := rune(0); r <= unicode.MaxRune; r++ {
		prop, _ := bidi.LookupRune(r)
		if !prop.IsBracket() {
			continue
		}
		pair := GetBracket(r)
		if pair.id() == NoBracket {
			t.Errorf("%U: missing bracket pair", r)
		}
		if pair.isOpen() != prop.IsOpeningBracket() {
			t.Errorf("%U: incorrect opening bracket flag", r)
		}
		if canonical, _ := bidi.LookupRune(rune(pair.id())); !canonical.IsBracket() {
			t.Errorf("%U: canonical pair %U is not a bracket", r, rune(pair.id()))
		}
	}
}

func TestBracketPairs(t *testing.T) {
	t.Parallel()
	for _, pair := range []struct{ open, close, id rune }{
		{'(', ')', '('},
		{'\u2329', '\u3009', '\u3008'},
		{'\u2e55', '\u2e56', '\u2e55'},
		{'\u2e57', '\u2e58', '\u2e57'},
		{'\u2e59', '\u2e5a', '\u2e59'},
		{'\u2e5b', '\u2e5c', '\u2e5b'},
	} {
		t.Run(fmt.Sprintf("%U/%U", pair.open, pair.close), func(t *testing.T) {
			if prop, _ := bidi.LookupRune(pair.open); !prop.IsBracket() {
				t.Skipf("%U is not a bracket in Unicode %s", pair.open, bidi.UnicodeVersion)
			}
			if got := GetBracket(pair.open); got != bracketOpenMask|BracketType(pair.id) {
				t.Errorf("GetBracket(%U) = %#x, want %#x", pair.open, got, bracketOpenMask|BracketType(pair.id))
			}
			if got := GetBracket(pair.close); got != BracketType(pair.id) {
				t.Errorf("GetBracket(%U) = %#x, want %#x", pair.close, got, pair.id)
			}

			var base ParType = ON
			text := []rune(fmt.Sprintf("א %ca%c aa", pair.open, pair.close))
			visual, _ := LogicalToVisual(0, text, &base)
			if want := []Level{1, 1, 1, 2, 1, 1, 2, 2}; !reflect.DeepEqual(visual.EmbeddingLevels, want) {
				t.Errorf("embedding levels = %v, want %v", visual.EmbeddingLevels, want)
			}
			if want := []int{6, 7, 5, 4, 3, 2, 1, 0}; !reflect.DeepEqual(visual.VisualToLogical, want) {
				t.Errorf("visual order = %v, want %v", visual.VisualToLogical, want)
			}
		})
	}
}

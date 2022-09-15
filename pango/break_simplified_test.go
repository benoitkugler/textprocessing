package pango

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
)

func TestUCD2(t *testing.T) {
	files := [...]string{
		// "test/breaks/GraphemeBreakTest.txt",
		// "test/breaks/EmojiBreakTest.txt",
		// "test/breaks/CharBreakTest.txt",
		// "test/breaks/WordBreakTest.txt",
		// "test/breaks/SentenceBreakTest.txt",
		"test/breaks/LineBreakTest.txt",
	}
	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(string(b), "\n")
		for _, line := range lines {
			if len(line) == 0 || strings.HasPrefix(line, "#") {
				continue
			}
			s, _, err := parseLine(line)
			if err != nil {
				t.Fatal(err)
			}
			text := []rune(s)
			exp := make([]CharAttr, len(text)+1)
			pangoDefaultBreak(text, exp)

			got := make([]CharAttr, len(text)+1)
			pangoDefaultBreak2(text, got)

			assertEqualAttrs(t, got, exp)
		}
	}
}

func parseLine(line string) (string, []bool, error) {
	var attrReturn []bool
	var gs string

	line = strings.Split(line, "#")[0]
	for _, field := range strings.Fields(line) {
		switch field {
		case string(rune(0x00f7)): /* DIVISION SIGN: boundary here */
			attrReturn = append(attrReturn, true)
		case string(rune(0x00d7)): /* MULTIPLICATION SIGN: no boundary here */
			attrReturn = append(attrReturn, false)
		default:
			character, err := strconv.ParseUint(field, 16, 32)
			if err != nil {
				return "", nil, fmt.Errorf("invalid line %s: %s", line, err)
			}
			if character > 0x10ffff {
				return "", nil, fmt.Errorf("unexpected character")
			}
			gs += string(rune(character))
		}
	}
	return gs, attrReturn, nil
}

func assertEqualAttrs(t *testing.T, got, exp []CharAttr) {
	if len(got) != len(exp) {
		t.Fatalf("exepected length %d, got %d", len(got), len(exp))
	}
	for i := range got {
		const mask = LineBreak | MandatoryBreak
		if (mask & got[i]) != (mask & exp[i]) {
			t.Errorf("wrong value at index %d", i)
		}
	}
}

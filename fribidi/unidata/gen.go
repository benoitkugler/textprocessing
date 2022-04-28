package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/benoitkugler/textprocessing/fribidi"
)

/* We do not support surrogates yet */
const maxUnicode = 0x110000

// read UnicodeData to build a Go lookup table

var (
	shapingTable struct {
		table    [maxUnicode][4]rune
		min, max rune
	}
	equivTable [maxUnicode]rune
)

func parseUnicodeData() error {
	// initialisation
	for c := range shapingTable.table {
		for i := range shapingTable.table[c] {
			shapingTable.table[c][i] = rune(c)
		}
	}

	b, err := ioutil.ReadFile("UnicodeData.txt")
	if err != nil {
		return err
	}
	var (
		min rune = maxUnicode
		max rune
	)
	for _, l := range bytes.Split(b, []byte{'\n'}) {
		line := string(bytes.TrimSpace(l))
		if line == "" || line[0] == '#' { // reading header or comment
			continue
		}
		chunks := strings.Split(line, ";")

		// rune;comment;type general;...;type bidi;...;<...> XXX

		// we are looking for <...> XXXX
		if len(chunks) < 6 || chunks[5] == "" {
			continue
		}
		var (
			c        rune
			tag      string
			unshaped rune
		)

		_, err = fmt.Sscanf(chunks[0], "%04x", &c)
		if err != nil {
			return fmt.Errorf("invalid line %s: %s", line, err)
		}
		if c >= maxUnicode || unshaped >= maxUnicode {
			return fmt.Errorf("invalid line %s: too high rune value", line)
		}

		if chunks[5][0] == '<' {
			_, err = fmt.Sscanf(chunks[5], "%s %04x", &tag, &unshaped)
		} else {
			_, err = fmt.Sscanf(chunks[5], "%04x", &unshaped)
		}
		if err != nil {
			return fmt.Errorf("invalid line %s: %s", line, err)
		}

		// shape table: only single unshaped rune are considered
		if shape := isShape(tag); shape >= 0 && len(chunks[5]) == len(tag)+5 {
			shapingTable.table[unshaped][shape] = c
			if unshaped < min {
				min = unshaped
			}
			if unshaped > max {
				max = unshaped
			}
		}

		// equiv table
		equivTable[c] = unshaped
	}
	shapingTable.min, shapingTable.max = min, max

	return nil
}

func isShape(s string) int {
	for i, tag := range [...]string{
		"<isolated>",
		"<final>",
		"<initial>",
		"<medial>",
	} {
		if tag == s {
			return i
		}
	}
	return -1
}

func parseBrackets() (map[rune]rune, error) {
	out := map[rune]rune{}

	b, err := ioutil.ReadFile("BidiBrackets.txt")
	if err != nil {
		return nil, err
	}

	for _, l := range bytes.Split(b, []byte{'\n'}) {
		line := string(bytes.TrimSpace(l))
		if line == "" || line[0] == '#' { // reading header or comment
			continue
		}
		var (
			i, j        rune
			openOrClose string
		)
		_, err := fmt.Sscanf(line, "%04x; %04x; %s ", &i, &j, &openOrClose)
		if err != nil {
			return nil, fmt.Errorf("invalid line %s: %s", line, err)
		}

		if i >= maxUnicode || j >= maxUnicode {
			return nil, fmt.Errorf("to high rune value: %d and %d", i, j)
		}

		// Open braces map to themself
		if openOrClose == "o" {
			j = i
		}

		// Turn j into the unicode equivalence if it exists
		if equivTable[j] != 0 {
			j = equivTable[j]
		}

		out[i] = j
	}
	return out, nil
}

func parseArabicJoining() (map[rune]fribidi.JoiningType, error) {
	b, err := ioutil.ReadFile("ArabicShaping.txt")
	if err != nil {
		return nil, err
	}

	out := make(map[rune]fribidi.JoiningType)

	for _, l := range bytes.Split(b, []byte{'\n'}) {
		line := string(bytes.TrimSpace(l))
		if line == "" || line[0] == '#' { // reading header or comment
			continue
		}

		fields := strings.Split(line, ";")
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid line %s", line)
		}

		var c rune
		_, err := fmt.Sscanf(fields[0], "%x", &c)
		if err != nil {
			return nil, fmt.Errorf("invalid line %s: %s", line, err)
		}

		if c >= maxUnicode {
			return nil, fmt.Errorf("to high rune value: %d", c)
		}

		j, err := getJoiningType(strings.TrimSpace(fields[2]))
		if err != nil {
			return nil, err
		}

		out[c] = j
	}
	return out, nil
}

var joiningTypes = [...]string{
	fribidi.U: "U",
	fribidi.R: "R",
	fribidi.D: "D",
	fribidi.C: "C",
	fribidi.T: "T",
	fribidi.L: "L",
	fribidi.G: "G",
}

func getJoiningType(s string) (fribidi.JoiningType, error) {
	for j, v := range joiningTypes {
		if v == s {
			return fribidi.JoiningType(j), nil
		}
	}
	return 0, fmt.Errorf("unknown join type: %s", s)
}

// ---------------------------- export function ----------------------------

const accesFunc = `
func getArabicShapePres(r rune, shape uint8) rune {
	if r < %d || r > %d {
		return r
	}
	return %s[r-%d][shape]
}
`

func genShapingTable(out io.Writer) error {
	if shapingTable.max < shapingTable.min {
		return errors.New("error: no shaping pair found, something wrong with reading input")
	}

	const tableName = "arShap"

	_, err := fmt.Fprintln(out, "package fribidi\n\n // Code generated by fribidi/unidata/gen.go from UnicodeData.txt DO NOT EDIT")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "// required memory: %d KB\n", (shapingTable.max-shapingTable.min+1)*4*4/1000)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "var %s = [...][4]rune{\n", tableName)
	if err != nil {
		return err
	}
	for c := shapingTable.min; c <= shapingTable.max; c++ {
		_, err = fmt.Fprintf(out, "%#v,\n", shapingTable.table[c])
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(out, "}\n\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, accesFunc, shapingTable.min, shapingTable.max, tableName, shapingTable.min)
	return err
}

func genBracketsTable(m map[rune]rune, out io.Writer) error {
	_, err := fmt.Fprintln(out, "package fribidi\n\n // Code generated by fribidi/unidata/gen.go from UnicodeData.txt DO NOT EDIT")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "// map length: %d\n", len(m))
	if err != nil {
		return err
	}

	// sort keys to have deterministic output
	var keys []rune
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	s := "var bracketsTable = map[rune]rune{\n"
	for _, k := range keys {
		s += fmt.Sprintf("0x%x:%d,\n", k, m[k])
	}
	s += "}"

	_, err = fmt.Fprintf(out, s)
	return err
}

func genJoiningTable(m map[rune]fribidi.JoiningType, out io.Writer) error {
	_, err := fmt.Fprintln(out, "package fribidi\n\n // Code generated by fribidi/unidata/gen.go from UnicodeData.txt DO NOT EDIT")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "// map length: %d\n", len(m))
	if err != nil {
		return err
	}

	// sort keys to have deterministic output
	var keys []rune
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	s := "var joiningTable = map[rune]JoiningType{\n"
	for _, k := range keys {
		s += fmt.Sprintf("0x%x:%d,\n", k, m[k])
	}
	s += "}"

	_, err = fmt.Fprint(out, s)
	return err
}

func main() {
	err := parseUnicodeData()
	if err != nil {
		log.Fatal(err)
	}

	fileShaping := "../arabic_table.go"
	f, err := os.Create(fileShaping)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	err = genShapingTable(f)
	if err != nil {
		log.Fatal(err)
	}
	err = exec.Command("goimports", "-w", fileShaping).Run()
	if err != nil {
		log.Fatal("can't format: ", err)
	}

	m, err := parseBrackets()
	if err != nil {
		log.Fatal(err)
	}
	fileBrackets := "../brackets_table.go"
	f, err = os.Create(fileBrackets)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	err = genBracketsTable(m, f)
	if err != nil {
		log.Fatal(err)
	}
	err = exec.Command("goimports", "-w", fileBrackets).Run()
	if err != nil {
		log.Fatal("can't format: ", err)
	}

	m2, err := parseArabicJoining()
	if err != nil {
		log.Fatal(err)
	}
	fileJoining := "../joining_table.go"
	f, err = os.Create(fileJoining)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	err = genJoiningTable(m2, f)
	if err != nil {
		log.Fatal(err)
	}
	err = exec.Command("goimports", "-w", fileJoining).Run()
	if err != nil {
		log.Fatal("can't format: ", err)
	}

	fmt.Println("Done.")
}

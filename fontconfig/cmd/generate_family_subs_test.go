package main

import (
	"fmt"
	"testing"

	"github.com/benoitkugler/textprocessing/fontconfig"
)

func TestGenerate(t *testing.T) {
	substitutions, err := fontconfig.GenerateSubstitution()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(substitutions)
}

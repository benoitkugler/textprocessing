package fontconfig

import (
	"fmt"
	"log"
	"strings"

	"github.com/benoitkugler/textlayout/language"
)

func exprAsString(expr *expression) string {
	if s, ok := expr.u.(String); ok && expr.op == opString {
		return string(s)
	}
	return ""
}

func (d directive) editFamily() []ruleEdit {
	var out []ruleEdit
	for _, edit := range d.edits {
		if edit.object == FAMILY {
			out = append(out, edit)
		}
	}
	return out
}

func exprAsStringList(expr *expression) []string {
	if s, ok := expr.u.(String); ok && expr.op == opString {
		return []string{string(s)}
	}

	if c, ok := expr.u.(exprTree); ok && expr.op == opComma {
		l1 := exprAsStringList(c.left)
		l2 := exprAsStringList(c.right)
		return append(l1, l2...)
	}

	return nil
}

type ExportedFamilySubstitution struct {
	Comment            string
	TestCode           string
	OpCode             string   // how to insert the families
	AdditionalFamilies []string // the families to add
	Importance         byte     // how is the precedence
}

// GenerateSubstitution exports the Standard family substitution
// rules
func GenerateSubstitution() ([]ExportedFamilySubstitution, error) {
	var substitutions []ExportedFamilySubstitution

	for _, ruleset := range Standard.subst {
		comment := ruleset.name
		if ruleset.description != "" {
			comment += fmt.Sprintf(" (%s)", ruleset.description)
		}

		for _, directive := range ruleset.subst[MatchQuery] {
			edits := directive.editFamily()
			if len(edits) == 0 {
				continue
			}
			if len(edits) != 1 {
				return nil, fmt.Errorf("edit not supported: %v", edits)
			}
			edit := edits[0]
			subs := ExportedFamilySubstitution{
				Comment:            comment,
				AdditionalFamilies: exprAsStringList(edit.expr),
			}

			switch edit.op.getOp() {
			case opAppend:
				subs.OpCode = "opAppend"
			case opAppendLast:
				subs.OpCode = "opAppendLast"
			case opPrepend:
				subs.OpCode = "opPrepend"
			case opPrependFirst:
				subs.OpCode = "opPrependFirst"
			case opAssign:
				subs.OpCode = "opReplace"
			default:
				return nil, fmt.Errorf("unexpected operation %v", edit.op)
			}

			// extract the test

			tests := directive.tests
			switch len(tests) {
			case 1:
				test := tests[0]
				if test.object == SPACING {
					log.Println("ignored test", tests)
					continue
				}

				if test.object != FAMILY {
					return nil, fmt.Errorf("test not supported: %v", test)
				}

				familyTarget := exprAsString(test.expr)
				if familyTarget == "" {
					return nil, fmt.Errorf("test not supported: %v", test)
				}

				switch test.op.getOp() {
				case opEqual:
					subs.TestCode = fmt.Sprintf("familyEquals(%q)", familyTarget)
				case opContains:
					subs.TestCode = fmt.Sprintf("familyContains(%q)", familyTarget)
				default:
					return nil, fmt.Errorf("test not supported: %v", test)
				}
			case 2:
				var (
					lang, fam     string
					langOp, famOp opKind
				)
				if tests[0].object == FAMILY && tests[1].object == LANG {
					lang, fam = exprAsString(tests[1].expr), exprAsString(tests[0].expr)
					langOp, famOp = tests[1].op.getOp(), tests[0].op.getOp()
				} else if tests[1].object == FAMILY && tests[0].object == LANG {
					lang, fam = exprAsString(tests[0].expr), exprAsString(tests[1].expr)
					langOp, famOp = tests[0].op.getOp(), tests[1].op.getOp()
				} else {
					log.Println("ignored test", tests)
					continue
				}

				langVar := "language.Lang" + strings.ReplaceAll(strings.Title(string(language.NewLanguage(lang))), "-", "_")

				switch langOp {
				case opEqual:
					switch famOp {
					case opEqual:
						subs.TestCode = fmt.Sprintf("langAndFamilyEqual{lang: %s,family: %q}", langVar, fam)
					case opNotEqual:
						subs.TestCode = fmt.Sprintf("langEqualsAndNoFamily{lang: %s,family: %q}", langVar, fam)
						subs.OpCode = "opAppendLast"
					default:
						return nil, fmt.Errorf("family op not supported: %v", tests)
					}
				case opContains:
					if famOp != opEqual {
						return nil, fmt.Errorf("family op not supported: %v", tests)
					}
					// we replace contains by equal since lang are already normalized
					subs.TestCode = fmt.Sprintf("langAndFamilyEqual{lang: %s,family: %q}", langVar, fam)
				default:
					return nil, fmt.Errorf("test not supported: %v", tests)
				}
			case 3:
				// we special case the generic fallback
				if tests[0].object == FAMILY && tests[1].object == FAMILY && tests[2].object == FAMILY {
					subs.TestCode = "noGenericFamily{}"
				} else {
					log.Println("ignored test", tests)
					continue
				}
			default:
				log.Println("ignored test", tests)
				continue
			}

			// add and check the precedence
			switch edit.binding {
			case vbWeak:
				subs.Importance = 'w'
			case vbSame:
				if !(strings.HasPrefix(subs.TestCode, "familyEquals") || strings.HasPrefix(subs.TestCode, "langAndFamilyEqual") || strings.HasPrefix(subs.TestCode, "familyContains")) {
					return nil, fmt.Errorf("unsupported precedence 'equal' for test %s", subs.TestCode)
				}
				subs.Importance = 'e'
			case vbStrong:
				subs.Importance = 's'
			}

			substitutions = append(substitutions, subs)
		}
	}

	return substitutions, nil
}

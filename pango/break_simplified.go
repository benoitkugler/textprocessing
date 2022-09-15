package pango

import (
	"unicode"

	ucd "github.com/benoitkugler/textlayout/unicodedata"
)

// graphemeType returns the graphemeBreakType of the rune
func graphemeType(r rune, generalCategorie *unicode.RangeTable, previousType graphemeBreakType, makesHangulSyllable bool) graphemeBreakType {
	gbType := gbOther
	switch generalCategorie {
	case unicode.Cf:
		if r == 0x200C {
			gbType = gbExtend
			break
		} else if r == 0x200D {
			gbType = gbZWJ
			break
		} else if (r >= 0x600 && r <= 0x605) ||
			r == 0x6DD ||
			r == 0x70F ||
			r == 0x8E2 ||
			r == 0xD4E ||
			r == 0x110BD ||
			(r >= 0x111C2 && r <= 0x111C3) {
			gbType = gbPrepend
			break
		}
		/* Tag chars */
		if r >= 0xE0020 && r <= 0xE00FF {
			gbType = gbExtend
			break
		}
		fallthrough
	case unicode.Cc, unicode.Zl, unicode.Zp, unicode.Cs:
		gbType = gbControlCRLF
	case nil:
		/* Unassigned default ignorables */
		if (r >= 0xFFF0 && r <= 0xFFF8) || (r >= 0xE0000 && r <= 0xE0FFF) {
			gbType = gbControlCRLF
			break
		}
		fallthrough
	case unicode.Lo:
		if makesHangulSyllable {
			gbType = gbInHangulSyllable
		}
	case unicode.Lm:
		if r >= 0xFF9E && r <= 0xFF9F {
			gbType = gbExtend /* Other_Grapheme_Extend */
		}
	case unicode.Mc:
		gbType = gbSpacingMark /* SpacingMark */
		if r >= 0x0900 {
			if r == 0x09BE || r == 0x09D7 ||
				r == 0x0B3E || r == 0x0B57 || r == 0x0BBE || r == 0x0BD7 ||
				r == 0x0CC2 || r == 0x0CD5 || r == 0x0CD6 ||
				r == 0x0D3E || r == 0x0D57 || r == 0x0DCF || r == 0x0DDF ||
				r == 0x1D165 || (r >= 0x1D16E && r <= 0x1D172) {
				gbType = gbExtend /* Other_Grapheme_Extend */
			}
		}
	case unicode.Me, unicode.Mn:
		gbType = gbExtend /* Grapheme_Extend */
	case unicode.So:
		if r >= 0x1F1E6 && r <= 0x1F1FF {
			if previousType == gbRiOdd {
				gbType = gbRiEven
			} else {
				gbType = gbRiOdd
			}
		}
	case unicode.Sk:
		/* Fitzpatrick modifiers */
		if r >= 0x1F3FB && r <= 0x1F3FF {
			gbType = gbExtend
		}
	}
	return gbType
}

// newlineBreakType returns the LineBreakType of r
func newlineBreakType(breakType *unicode.RangeTable, prev lineBreakType) lineBreakType {
	switch breakType {
	case ucd.BreakNU:
		return lbNumeric
	case ucd.BreakSY, ucd.BreakIS:
		if prev != lbNumeric {
			return lbOther
		}
	case ucd.BreakCL, ucd.BreakCP:
		if prev == lbNumeric {
			return lbNumericClose
		}
		return lbOther
	case ucd.BreakRI:
		if prev == lbRiOdd {
			return lbRiEven
		}
		return lbRiOdd
	}
	return lbOther
}

func ruleLB30(breakOp *breakOpportunity, prevBreakType, breakType *unicode.RangeTable) {
	if (prevBreakType == ucd.BreakAL ||
		prevBreakType == ucd.BreakHL ||
		prevBreakType == ucd.BreakNU) &&
		breakType == ucd.BreakOP {
		*breakOp = breakProhibited
	}
	if prevBreakType == ucd.BreakCP &&
		(breakType == ucd.BreakAL ||
			breakType == ucd.BreakHL ||
			breakType == ucd.BreakNU) {
		*breakOp = breakProhibited
	}
}

func ruleLB30ab(breakOp *breakOpportunity, prevBreakType, breakType *unicode.RangeTable, prev rune,
	lbType, prevLbType lineBreakType,
) {
	/* Rule LB30a */
	if prevLbType == lbRiOdd && lbType == lbRiEven {
		*breakOp = breakProhibited
	}
	/* Rule LB30b */
	if prevBreakType == ucd.BreakEB && breakType == ucd.BreakEM {
		*breakOp = breakProhibited
	}

	if unicode.Is(ucd.Extended_Pictographic, prev) && ucd.LookupType(prev) == nil &&
		breakType == ucd.BreakEM {
		*breakOp = breakProhibited
	}
}

func ruleLB29To26(breakOp *breakOpportunity, prevBreakType, breakType *unicode.RangeTable) {
	/* Rule LB29 */
	if prevBreakType == ucd.BreakIS && (breakType == ucd.BreakAL ||
		breakType == ucd.BreakHL) {
		*breakOp = breakProhibited
	}
	/* Rule LB28 */
	if (prevBreakType == ucd.BreakAL || prevBreakType == ucd.BreakHL) &&
		(breakType == ucd.BreakAL || breakType == ucd.BreakHL) {
		*breakOp = breakProhibited
	}
	/* Rule LB27 */
	if (prevBreakType == ucd.BreakJL ||
		prevBreakType == ucd.BreakJV ||
		prevBreakType == ucd.BreakJT ||
		prevBreakType == ucd.BreakH2 ||
		prevBreakType == ucd.BreakH3) && breakType == ucd.BreakPO {
		*breakOp = breakProhibited
	}
	if prevBreakType == ucd.BreakPR &&
		(breakType == ucd.BreakJL ||
			breakType == ucd.BreakJV ||
			breakType == ucd.BreakJT ||
			breakType == ucd.BreakH2 ||
			breakType == ucd.BreakH3) {
		*breakOp = breakProhibited
	}
	/* Rule LB26 */
	if prevBreakType == ucd.BreakJL &&
		(breakType == ucd.BreakJL ||
			breakType == ucd.BreakJV ||
			breakType == ucd.BreakH2 ||
			breakType == ucd.BreakH3) {
		*breakOp = breakProhibited
	}
	if (prevBreakType == ucd.BreakJV || prevBreakType == ucd.BreakH2) &&
		(breakType == ucd.BreakJV || breakType == ucd.BreakJT) {
		*breakOp = breakProhibited
	}
	if (prevBreakType == ucd.BreakJT || prevBreakType == ucd.BreakH3) &&
		breakType == ucd.BreakJT {
		*breakOp = breakProhibited
	}
}

func ruleLB25(breakOp *breakOpportunity, prevBreakType, breakType, nextBreakType *unicode.RangeTable, prevLbType lineBreakType) {
	/* Rule LB25 with Example 7 of Customization */
	if (prevBreakType == ucd.BreakPR || prevBreakType == ucd.BreakPO) &&
		breakType == ucd.BreakNU {
		*breakOp = breakProhibited
	}
	if (prevBreakType == ucd.BreakPR || prevBreakType == ucd.BreakPO) &&
		(breakType == ucd.BreakOP || breakType == ucd.BreakHY) &&
		nextBreakType == ucd.BreakNU {
		*breakOp = breakProhibited
	}
	if (prevBreakType == ucd.BreakOP || prevBreakType == ucd.BreakHY) &&
		breakType == ucd.BreakNU {
		*breakOp = breakProhibited
	}
	if prevBreakType == ucd.BreakNU &&
		(breakType == ucd.BreakNU ||
			breakType == ucd.BreakSY ||
			breakType == ucd.BreakIS) {
		*breakOp = breakProhibited
	}
	if prevLbType == lbNumeric &&
		(breakType == ucd.BreakNU ||
			breakType == ucd.BreakSY ||
			breakType == ucd.BreakIS ||
			breakType == ucd.BreakCL ||
			breakType == ucd.BreakCP) {
		*breakOp = breakProhibited
	}
	if (prevLbType == lbNumeric || prevLbType == lbNumericClose) &&
		(breakType == ucd.BreakPO || breakType == ucd.BreakPR) {
		*breakOp = breakProhibited
	}
}

func ruleLB24To22(breakOp *breakOpportunity, prevBreakType, breakType *unicode.RangeTable) {
	/* Rule LB24 */
	if (prevBreakType == ucd.BreakPR ||
		prevBreakType == ucd.BreakPO) &&
		(breakType == ucd.BreakAL ||
			breakType == ucd.BreakHL) {
		*breakOp = breakProhibited
	}
	if (prevBreakType == ucd.BreakAL ||
		prevBreakType == ucd.BreakHL) &&
		(breakType == ucd.BreakPR || breakType == ucd.BreakPO) {
		*breakOp = breakProhibited
	}
	/* Rule LB23 */
	if (prevBreakType == ucd.BreakAL ||
		prevBreakType == ucd.BreakHL) &&
		breakType == ucd.BreakNU {
		*breakOp = breakProhibited
	}
	if prevBreakType == ucd.BreakNU &&
		(breakType == ucd.BreakAL ||
			breakType == ucd.BreakHL) {
		*breakOp = breakProhibited
	}
	/* Rule LB23a */
	if prevBreakType == ucd.BreakPR &&
		(breakType == ucd.BreakID ||
			breakType == ucd.BreakEB ||
			breakType == ucd.BreakEM) {
		*breakOp = breakProhibited
	}
	if (prevBreakType == ucd.BreakID ||
		prevBreakType == ucd.BreakEB ||
		prevBreakType == ucd.BreakEM) &&
		breakType == ucd.BreakPO {
		*breakOp = breakProhibited
	}

	/* Rule LB22 */
	if breakType == ucd.BreakIN {
		if prevBreakType == ucd.BreakAL ||
			prevBreakType == ucd.BreakHL {
			*breakOp = breakProhibited
		}
		if prevBreakType == ucd.BreakEX {
			*breakOp = breakProhibited
		}
		if prevBreakType == ucd.BreakID ||
			prevBreakType == ucd.BreakEB ||
			prevBreakType == ucd.BreakEM {
			*breakOp = breakProhibited
		}
		if prevBreakType == ucd.BreakIN {
			*breakOp = breakProhibited
		}
		if prevBreakType == ucd.BreakNU {
			*breakOp = breakProhibited
		}
	}
}

func ruleLB21To9(breakOp *breakOpportunity, prevPrevBreakType, prevBreakType, breakType, rowBreakType *unicode.RangeTable) {
	if breakType == ucd.BreakBA ||
		breakType == ucd.BreakHY ||
		breakType == ucd.BreakNS ||
		prevBreakType == ucd.BreakBB {
		*breakOp = breakProhibited /* Rule LB21 */
	}
	if prevPrevBreakType == ucd.BreakHL &&
		(prevBreakType == ucd.BreakHY ||
			prevBreakType == ucd.BreakBA) {
		*breakOp = breakProhibited /* Rule LB21a */
	}
	if prevBreakType == ucd.BreakSY &&
		breakType == ucd.BreakHL {
		*breakOp = breakProhibited /* Rule LB21b */
	}

	if prevBreakType == ucd.BreakCB ||
		breakType == ucd.BreakCB {
		*breakOp = breakAllowed /* Rule LB20 */
	}
	if prevBreakType == ucd.BreakQU ||
		breakType == ucd.BreakQU {
		*breakOp = breakProhibited /* Rule LB19 */
	}

	/* handle related rules for Space as state machine here,
	   and override the pair table result. */
	if prevBreakType == ucd.BreakSP { /* Rule LB18 */
		*breakOp = breakAllowed
	}
	if rowBreakType == ucd.BreakB2 &&
		breakType == ucd.BreakB2 {
		*breakOp = breakProhibited /* Rule LB17 */
	}
	if (rowBreakType == ucd.BreakCL ||
		rowBreakType == ucd.BreakCP) &&
		breakType == ucd.BreakNS {
		*breakOp = breakProhibited /* Rule LB16 */
	}
	if rowBreakType == ucd.BreakQU &&
		breakType == ucd.BreakOP {
		*breakOp = breakProhibited /* Rule LB15 */
	}
	if rowBreakType == ucd.BreakOP {
		*breakOp = breakProhibited /* Rule LB14 */
	}
	/* Rule LB13 with Example 7 of Customization */
	if breakType == ucd.BreakEX {
		*breakOp = breakProhibited
	}
	if prevBreakType != ucd.BreakNU &&
		(breakType == ucd.BreakCL ||
			breakType == ucd.BreakCP ||
			breakType == ucd.BreakIS ||
			breakType == ucd.BreakSY) {
		*breakOp = breakProhibited
	}
	if prevBreakType == ucd.BreakGL {
		*breakOp = breakProhibited /* Rule LB12 */
	}
	if breakType == ucd.BreakGL &&
		(prevBreakType != ucd.BreakSP &&
			prevBreakType != ucd.BreakBA &&
			prevBreakType != ucd.BreakHY) {
		*breakOp = breakProhibited /* Rule LB12a */
	}
	if prevBreakType == ucd.BreakWJ ||
		breakType == ucd.BreakWJ {
		*breakOp = breakProhibited /* Rule LB11 */
	}

	/* Rule LB9 */
	if breakType == ucd.BreakCM ||
		breakType == ucd.BreakZWJ {
		if !(prevBreakType == ucd.BreakBK ||
			prevBreakType == ucd.BreakCR ||
			prevBreakType == ucd.BreakLF ||
			prevBreakType == ucd.BreakNL ||
			prevBreakType == ucd.BreakSP ||
			prevBreakType == ucd.BreakZW) {
			*breakOp = breakProhibited
		}
	}
}

func ruleLB8(breakOp *breakOpportunity, prev rune, rowBreakType *unicode.RangeTable) {
	if rowBreakType == ucd.BreakZW {
		*breakOp = breakAllowed /* Rule LB8 */
	}
	if prev == 0x200D {
		*breakOp = breakProhibited /* Rule LB8a */
	}
}

func ruleLB7To6(breakOp *breakOpportunity, breakType *unicode.RangeTable) {
	if breakType == ucd.BreakSP ||
		breakType == ucd.BreakZW {
		*breakOp = breakProhibited /* Rule LB7 */
	}
	/* Rule LB6 */
	if breakType == ucd.BreakBK ||
		breakType == ucd.BreakCR ||
		breakType == ucd.BreakLF ||
		breakType == ucd.BreakNL {
		*breakOp = breakProhibited
	}
}

func ruleLB5To4(breakOp *breakOpportunity, r rune, prevBreakType *unicode.RangeTable) (mandatoryBreak bool) {
	/* Rules LB4 and LB5 */
	if prevBreakType == ucd.BreakBK || (prevBreakType == ucd.BreakCR && r != '\n') ||
		prevBreakType == ucd.BreakLF || prevBreakType == ucd.BreakNL {
		*breakOp = breakAllowed
		return true
	}
	return false
}

func ruleLB1(breakType, generalCategory *unicode.RangeTable) *unicode.RangeTable {
	switch breakType {
	case ucd.BreakAI, ucd.BreakSG, ucd.BreakXX:
		return ucd.BreakAL
	case ucd.BreakSA:
		if generalCategory == unicode.Mn || generalCategory == unicode.Mc {
			return ucd.BreakCM
		} else {
			return ucd.BreakAL
		}
	case ucd.BreakCJ:
		return ucd.BreakNS
	}
	return breakType
}

func ruleGB11(prevGbType, gbType graphemeBreakType, prev rune, isExtendedPictographic bool) bool {
	if gbType == gbExtend {
		return true
	} else if unicode.Is(ucd.Extended_Pictographic, prev) && gbType == gbZWJ {
		return true
	} else if prevGbType == gbExtend && gbType == gbZWJ {
		return true
	} else if prevGbType == gbZWJ && isExtendedPictographic {
		return true
	}
	return false
}

// apply the Grapheme Cluster Boundary Rules
func isGraphemeBoundary(prev, r rune, prevGbType, gbType graphemeBreakType, isExtendedPictographic, metExtendedPictographic bool) bool {
	if r == '\n' && prev == '\r' {
		return false /* Rule GB3 */
	} else if prevGbType == gbControlCRLF || gbType == gbControlCRLF {
		return true /* Rules GB4 && GB5 */
	} else if gbType == gbInHangulSyllable {
		return false /* Rules GB6, GB7, GB8 */
	} else if gbType == gbExtend {
		return false /* Rule GB9 */
	} else if gbType == gbZWJ {
		return false /* Rule GB9 */
	} else if gbType == gbSpacingMark {
		return false /* Rule GB9a */
	} else if prevGbType == gbPrepend {
		return false /* Rule GB9b */
	} else if isExtendedPictographic { /* Rule GB11 */
		if prevGbType == gbZWJ && metExtendedPictographic {
			return false
		}
	} else if prevGbType == gbRiOdd && gbType == gbRiEven {
		return false /* Rule GB12 && GB13 */
	}
	return true // Rule GB999
}

// determine wheter this forms a Hangul syllable with the previous rune.
func isHangul(prevJamo, jamo ucd.JamoType) bool {
	if jamo == ucd.NO_JAMO {
		return false
	}
	prevEnd := ucd.HangulJamoProps[prevJamo].End
	thisStart := ucd.HangulJamoProps[jamo].Start
	return (prevEnd == thisStart) || (prevEnd+1 == thisStart)
}

// returns the new prevPrevBreakType, prevBreakType and prevJamo
func adjustBreakTypes(prevJamo, jamo ucd.JamoType, prevPrevBreakType, prevBreakType, breakType *unicode.RangeTable, isStart bool) (*unicode.RangeTable, *unicode.RangeTable, ucd.JamoType) {
	if breakType != ucd.BreakSP {
		/* Rule LB9 */
		if breakType == ucd.BreakCM || breakType == ucd.BreakZWJ {
			if isStart ||
				prevBreakType == ucd.BreakBK ||
				prevBreakType == ucd.BreakCR ||
				prevBreakType == ucd.BreakLF ||
				prevBreakType == ucd.BreakNL ||
				prevBreakType == ucd.BreakSP ||
				prevBreakType == ucd.BreakZW {
				// Rule LB10
				return prevPrevBreakType, ucd.BreakAL, jamo
			}
			// else don't change the prevBreakType for Rule LB9
			return prevPrevBreakType, prevBreakType, jamo
		} else {
			return prevBreakType, breakType, jamo
		}
	}

	if prevBreakType != ucd.BreakSP {
		return prevBreakType, breakType, prevJamo
	}
	// else don't change the prevBreakType
	return prevPrevBreakType, prevBreakType, prevJamo
}

// This is the default break algorithm. It applies Unicode
// rules without language-specific tailoring.
// To avoid allocations, `attrs` must be passed as input, and must have a length of len(text)+1.
func pangoDefaultBreak2(text []rune, attrs []CharAttr) {
	// many of the decisions rely on the comparison between the current
	// rune and the previous one
	var (
		prevWc rune

		prevJamo = ucd.NO_JAMO

		prevBreakType     *unicode.RangeTable
		prevPrevBreakType = ucd.BreakXX

		prevGbType              = gbOther
		metExtendedPictographic = false

		prevLbType = lbOther
	)

	nextWc := paragraphSeparator
	if len(text) != 0 {
		nextWc = text[0]
	}
	nextBreakType := ucd.LookupBreakClass(nextWc)

	for i := 0; i <= len(text); i++ {
		wc := paragraphSeparator
		if i < len(text) {
			wc = text[i]
		}

		breakType := nextBreakType // avoid calling LookupBreakClass twice
		if i == len(text) {
			nextWc = 0
		} else if i == len(text)-1 {
			// we fill in the last element of `attrs` by assuming there's a paragraph separators off the end
			// of text
			nextWc = paragraphSeparator
		} else {
			nextWc = text[i+1]
		}
		nextBreakType = ucd.LookupBreakClass(nextWc)

		// query general unicode properties for the current rune
		generalCategory := ucd.LookupType(wc)
		jamo := ucd.Jamo(breakType)
		isExtendedPictographic := unicode.Is(ucd.Extended_Pictographic, wc)

		makesHangulSyllable := isHangul(prevJamo, jamo)

		// UAX#29 Grapheme Boundaries, required for line breaking

		gbType := graphemeType(wc, generalCategory, prevGbType, makesHangulSyllable) // find the GraphemeBreakType of wc

		if metExtendedPictographic { // Rule GB11
			metExtendedPictographic = ruleGB11(prevGbType, gbType, prevWc, isExtendedPictographic)
		}

		isGB := isGraphemeBoundary(prevWc, wc, prevGbType, gbType, isExtendedPictographic, metExtendedPictographic)

		// line breaking
		breakType = ruleLB1(breakType, generalCategory)

		rowBreakType := prevBreakType
		if prevBreakType == ucd.BreakSP {
			rowBreakType = prevPrevBreakType
		}
		if rowBreakType == ucd.BreakXX {
			rowBreakType = ucd.BreakAL
		}

		attrs[i].setLineBreak(false)
		attrs[i].setMandatoryBreak(false)
		// if it's not a grapheme boundary, it's not a line break either
		earlyCheck := isGB ||
			breakType == ucd.BreakEM ||
			breakType == ucd.BreakZWJ ||
			breakType == ucd.BreakCM ||
			breakType == ucd.BreakJL ||
			breakType == ucd.BreakJV ||
			breakType == ucd.BreakJT ||
			breakType == ucd.BreakH2 ||
			breakType == ucd.BreakH3 ||
			breakType == ucd.BreakRI

		if earlyCheck {
			lbType := newlineBreakType(breakType, prevLbType)

			attrs[i].setLineBreak(true) // Rule LB31

			// add the line break rules in reverse order to override
			// the lower priority rules.
			breakOp := breakAlreadyHandled
			ruleLB30(&breakOp, prevBreakType, breakType)
			ruleLB30ab(&breakOp, prevBreakType, breakType, prevWc, lbType, prevLbType)
			ruleLB29To26(&breakOp, prevBreakType, breakType)
			ruleLB25(&breakOp, prevBreakType, breakType, nextBreakType, prevLbType)
			ruleLB24To22(&breakOp, prevBreakType, breakType)
			ruleLB21To9(&breakOp, prevPrevBreakType, prevBreakType, breakType, rowBreakType)
			ruleLB8(&breakOp, prevWc, rowBreakType)
			ruleLB7To6(&breakOp, breakType)
			isMandatoryBreak := ruleLB5To4(&breakOp, wc, prevBreakType)

			if isMandatoryBreak {
				attrs[i].setMandatoryBreak(true)
			}

			switch breakOp {
			case breakProhibited: // can't break here
				attrs[i].setLineBreak(false)
			case breakIfSpaces: // break if prev char was space
				if prevBreakType != ucd.BreakSP {
					attrs[i].setLineBreak(false)
				}
			case breakAllowed:
				attrs[i].setLineBreak(true)
			}

			// rule LB9
			if !(breakType == ucd.BreakCM || breakType == ucd.BreakZWJ) {
				/* Rule LB25 with Example 7 of Customization */
				if breakType == ucd.BreakNU || breakType == ucd.BreakSY || breakType == ucd.BreakIS {
					if prevLbType != lbNumeric {
						prevLbType = lbType
					} /* else don't change the prevLbType */
				} else {
					prevLbType = lbType
				}
			}
			// else don't change the prevLbType for Rule LB9
		}

		prevPrevBreakType, prevBreakType, prevJamo = adjustBreakTypes(prevJamo, jamo, prevPrevBreakType,
			prevBreakType, breakType, i == 0)
		prevGbType = gbType
		prevWc = wc
		if isExtendedPictographic {
			metExtendedPictographic = true
		}
	}

	attrs[0].setLineBreak(false)             // Rule LB2
	attrs[len(text)].setLineBreak(true)      // Rule LB3
	attrs[len(text)].setMandatoryBreak(true) // Rule LB3
}

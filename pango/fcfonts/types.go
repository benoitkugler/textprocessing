package fcfonts

import (
	"github.com/benoitkugler/textprocessing/fontconfig"
	fc "github.com/benoitkugler/textprocessing/fontconfig"
	"github.com/benoitkugler/textprocessing/pango"
)

type coverage fc.Charset

// Convert the given `charset` into a new Coverage object.
func fromCharset(charset fc.Charset) pango.Coverage {
	return (*coverage)(&charset)
}

// Get returns true if the rune is covered
func (c *coverage) Get(index rune) bool { return (*fc.Charset)(c).HasChar(index) }

func (c *coverage) Set(index rune, covered bool) {
	if covered {
		(*fc.Charset)(c).AddChar(index)
	} else {
		(*fc.Charset)(c).DelChar(index)
	}
}

// Copy returns a deep copy of the coverage
func (c *coverage) Copy() pango.Coverage {
	if c == nil {
		return c
	}
	cs := (*fc.Charset)(c).Copy()
	return (*coverage)(&cs)
}

// decoder represents a decoder that an application provides
// for handling a font that is encoded in a custom way.
type decoder interface {
	// GetCharset returns a charset given a font that
	// includes a list of supported characters in the font.
	// The implementation must be fast because the method is called
	// separately for each character to determine Unicode coverage.
	GetCharset(font *Font) fc.Charset

	// GetGlyph returns a single glyph for a given Unicode code point.
	GetGlyph(font *Font, r rune) pango.Glyph
}

type fcFontKeyHash struct {
	pattern    string
	variations string
	matrix     pango.Matrix
	contextKey int
}

type fontHash map[fcFontKeyHash]*Font // (GHashFunc)pango_font_key_hash,  (GEqualFunc)pango_font_key_equal

func (m fontHash) lookup(p fcFontKey) *Font {
	key := fcFontKeyHash{
		pattern: p.pattern.Hash(), matrix: p.matrix,
		contextKey: p.contextKey, variations: p.variations,
	}
	return m[key]
}

func (m fontHash) insert(key fcFontKey, v *Font) {
	keyCopy := fcFontKeyHash{
		pattern: key.pattern.Hash(), matrix: key.matrix,
		contextKey: key.contextKey, variations: key.variations,
	}
	v.key = &key
	m[keyCopy] = v
}

func (m fontHash) remove(p fcFontKey) {
	key := fcFontKeyHash{
		pattern: p.pattern.Hash(), matrix: p.matrix,
		contextKey: p.contextKey, variations: p.variations,
	}
	delete(m, key)
}

type fontsetCache map[fontsetKey]*Fontset

func (m fontsetCache) lookup(p fontsetKey) *Fontset {
	p.desc = p.desc.AsHash()
	p.fontmap = nil
	return m[p]
}

func (m fontsetCache) insert(p fontsetKey, v *Fontset) {
	p.desc = p.desc.AsHash()
	p.fontmap = nil
	m[p] = v
}

func (m fontsetCache) remove(p fontsetKey) {
	p.desc = p.desc.AsHash()
	p.fontmap = nil
	delete(m, p)
}

type patternHash map[string]*cachedPattern

func (m patternHash) lookup(p fontconfig.Pattern) *cachedPattern { return m[p.Hash()] }

func (m patternHash) insert(p fontconfig.Pattern, pts *cachedPattern) { m[p.Hash()] = pts }

// ------------------------------------------------------------------------------------

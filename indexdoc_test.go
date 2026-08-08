package tela

import (
	"fmt"
	"testing"

	"github.com/deroproject/derohe/dvm"
	"github.com/stretchr/testify/assert"
)

// TestParseINDEXForDOCs_NoPanic covers the reader crash (a header value whose
// token contains "DOC drove an out-of-range read). GetINDEXInfo reaches this
// with no recover, so a malicious INDEX crashed any client that opened it.
func TestParseINDEXForDOCs_NoPanic(t *testing.T) {
	cases := map[string]string{
		// An INDEX literally named "DOCS" - a completely ordinary name. The
		// value token "DOCS" lands last, so the old code read two tokens past it.
		"header value named DOCS": `Function InitializePrivate() Uint64
10 STORE("var_header_name", "DOCS")
20 STORE("DOC1", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
1000 RETURN 0
End Function`,
		"header value starting DOC": `Function InitializePrivate() Uint64
10 STORE("var_header_description", "DOCument viewer")
20 STORE("DOC1", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
1000 RETURN 0
End Function`,
	}
	for name, code := range cases {
		t.Run(name, func(t *testing.T) {
			sc, _, err := dvm.ParseSmartContract(code)
			if err != nil {
				t.Skipf("fixture did not parse: %v", err)
			}
			// Must not panic, and must still find the one real DOC.
			got := parseINDEXForDOCs(sc)
			assert.Equal(t, 1, len(got), "should find exactly the one real DOC")
		})
	}
}

// TestParseINDEXForDOCs_NumericOrder covers the lexicographic sort: DOC10..DOC13
// sorted before DOC2, mis-ordering every INDEX with 10 or more documents.
func TestParseINDEXForDOCs_NumericOrder(t *testing.T) {
	code := "Function InitializePrivate() Uint64\n"
	for i := 1; i <= 13; i++ {
		// scid encodes the intended index so order is checkable
		scid := fmt.Sprintf("%064d", i)
		code += fmt.Sprintf("%d STORE(\"DOC%d\", \"%s\")\n", i*10, i, scid)
	}
	code += "1000 RETURN 0\nEnd Function"

	sc, _, err := dvm.ParseSmartContract(code)
	assert.NoError(t, err)

	got := parseINDEXForDOCs(sc)
	assert.Equal(t, 13, len(got))
	for i, scid := range got {
		want := fmt.Sprintf("%064d", i+1) // position i must hold DOC(i+1)
		assert.Equal(t, want, scid, "DOC at position %d out of order", i)
	}
}

// TestDocKeyNumber pins the precise key match that replaces the substring test.
func TestDocKeyNumber(t *testing.T) {
	yes := map[string]int{`"DOC1"`: 1, `"DOC10"`: 10, `"DOC999"`: 999, "DOC2": 2}
	for tok, n := range yes {
		got, ok := docKeyNumber(tok)
		assert.True(t, ok, "%q should be a DOC key", tok)
		assert.Equal(t, n, got)
	}
	no := []string{`"DOCS"`, `"DOC"`, `"DOCument"`, `"var_header_name"`, `"DOC1a"`, `""`, `"scid"`}
	for _, tok := range no {
		_, ok := docKeyNumber(tok)
		assert.False(t, ok, "%q should NOT be a DOC key", tok)
	}
}

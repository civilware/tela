package tela

import (
	"go/parser"
	"strconv"
	"strings"
	"testing"

	"github.com/deroproject/derohe/dvm"
	"github.com/stretchr/testify/assert"
)

// dvmEvalString evaluates a quoted string literal exactly as the DVM does at
// runtime: it is parsed as an expression and unquoted (dvm.go, *ast.BasicLit
// token.STRING -> strconv.Unquote). A literal that fails here panics the DVM.
func dvmEvalString(literal string) (string, error) {
	if _, err := parser.ParseExpr(literal); err != nil {
		return "", err
	}
	return strconv.Unquote(literal)
}

// TestFormatValueRoundTrips is the core of the fix: every header value, however
// hostile, must survive formatValue -> on-chain literal -> DVM evaluation
// unchanged. The trailing backslash is the O14 case that escaped its own
// closing quote under the old fmt.Sprintf("%q"-less) formatting.
func TestFormatValueRoundTrips(t *testing.T) {
	cases := []string{
		"index.html",
		"A normal description",
		`My App\`,            // O14 - trailing backslash
		`ends with a quote"`, // embedded quote
		`both \ and "`,
		"line1\nline2\ttabbed", // control chars
		"café ☕ 日本語",           // non-ASCII
		"",                     // empty
		`STORE("x","y")`,       // looks like code
	}
	for _, in := range cases {
		lit := formatValue(in)
		got, err := dvmEvalString(lit)
		assert.NoError(t, err, "value %q produced a literal the DVM would panic on: %s", in, lit)
		assert.Equal(t, in, got, "value %q did not round-trip (literal %s)", in, lit)
	}
}

// TestFormatValueASCIIUnchanged pins that ordinary values are byte-identical to
// the previous formatting, so nothing about a normal contract's code changes.
func TestFormatValueASCIIUnchanged(t *testing.T) {
	for _, in := range []string{"index.html", "My App", "telatomicswaps.tela", "2d7b1452f42652c4"} {
		assert.Equal(t, `"`+in+`"`, formatValue(in), "ordinary value %q must format as before", in)
	}
}

// TestParseHeadersProducesEvaluableCode proves a full INDEX built with hostile
// header values emits literals the DVM can evaluate.
//
// Parsing is not the test. The old formatting produced code that parsed and
// failed later, at evaluation, where the DVM unquotes each string literal and
// panics if it is malformed - so this asserts on the literals themselves.
func TestParseHeadersProducesEvaluableCode(t *testing.T) {
	idx := &INDEX{
		DURL:    "test.tela",
		DOCs:    []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		Headers: Headers{NameHdr: `My App\`, DescrHdr: `has "quotes" and \ slash`},
	}
	code, err := ParseHeaders(TELA_INDEX_1, idx)
	assert.NoError(t, err)

	sc, _, err := dvm.ParseSmartContract(code)
	assert.NoError(t, err, "generated contract did not parse")

	// Every string literal the contract stores must survive the DVM's own
	// evaluation. Under the old formatting the nameHdr and descrHdr literals
	// fail here with "invalid syntax", which is a panic at runtime.
	checked := 0
	for name, function := range sc.Functions {
		if name != DVM_FUNC_INIT_PRIVATE {
			continue
		}
		for number, line := range function.Lines {
			for _, token := range line {
				if !strings.HasPrefix(token, `"`) {
					continue
				}
				checked++
				_, uerr := strconv.Unquote(token)
				assert.NoError(t, uerr, "line %v literal %s would panic the DVM", number, token)
			}
		}
	}
	assert.NotZero(t, checked, "no string literals were checked")
}

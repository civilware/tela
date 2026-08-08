package tela

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInjectMODRefusesExistingFunction covers injecting a MOD whose function
// name the contract already defines.
//
// The MOD's code is appended to the contract as text while its functions are
// set on the parsed contract, so overwriting produced a contract carrying two
// definitions of the same function. That parses, the MOD's body silently
// replaces the contract's own, and the served code carries a dead duplicate.
//
// The shipped tags are listed here rather than read from the package Mods
// variable, which keeps this test independent of anything registered at runtime.
func TestInjectMODRefusesExistingFunction(t *testing.T) {
	// The collision is on DeleteVar, the second name vsoo declares, so a
	// refusal which injected as it went would have already set SetVar
	base := `Function Initialize() Uint64
10 RETURN 0
End Function

Function DeleteVar(k String) Uint64
10 RETURN 77
End Function`

	modSC, _, err := Mods.injectMOD("vsoo", base)
	assert.Error(t, err, "injectMOD overwrote a function the contract already defined")
	_, injected := modSC.Functions["SetVar"]
	assert.False(t, injected, "refused MOD should not leave SetVar on the returned contract")

	_, _, err = Mods.InjectMODs("vsoo", base)
	assert.Error(t, err, "InjectMODs overwrote a function the contract already defined")

	// Every MOD the package ships must still inject into the INDEX template
	shipped := []string{"vsoo", "vsooim", "vspubsu", "vspubow", "vspubim", "txdwd", "txdwa", "txto"}
	for _, tag := range shipped {
		_, _, err = Mods.InjectMODs(tag, TELA_INDEX_1)
		assert.NoError(t, err, "MOD %q no longer injects into TELA-INDEX-1: %s", tag, err)
	}

	// The transfers MODClass is Multi MOD, so its members combine with a
	// variable store MOD in one contract and must not collide with each other
	combined := NewModTag([]string{"vsoo", "txdwd", "txdwa", "txto"})
	_, _, err = Mods.InjectMODs(combined, TELA_INDEX_1)
	assert.NoError(t, err, "valid MOD combination %q no longer injects: %s", combined, err)
}

// TestInjectMODRefusesUndeclaredFunction covers a MOD whose code defines a name
// it does not declare.
//
// The whole code set is appended to the contract as text, not only the declared
// functions, so a name the code defines but does not declare lands in the
// contract just the same. Checking the declared names alone would leave the
// contract carrying two definitions of it.
//
// The MODs value here is local, so this test does not read or write the package
// Mods variable.
func TestInjectMODRefusesUndeclaredFunction(t *testing.T) {
	class := MODClass{Name: "Test class", Tag: "tc"}
	m := MODs{
		mods: []MOD{{
			Name:        "Test mod",
			Tag:         class.NewTag("one"),
			Description: "injection test",
			FunctionCode: func() string {
				return `Function Declared() Uint64
10 RETURN 0
End Function

Function Undeclared() Uint64
10 RETURN 0
End Function`
			},
			FunctionNames: []string{"Declared"},
		}},
		classes: []MODClass{class},
		index:   []int{1},
	}

	base := `Function Undeclared() Uint64
10 RETURN 77
End Function`

	_, _, err := m.injectMOD("tcone", base)
	assert.Error(t, err, "injectMOD appended a second definition of a function the contract already defined")

	// A declared name the MOD's code does not define is injected as an empty
	// function, so it overwrites the contract's own just the same
	m.mods[0].FunctionNames = []string{"Declared", "NotInTheCode"}
	base = `Function NotInTheCode() Uint64
10 RETURN 77
End Function`

	_, _, err = m.injectMOD("tcone", base)
	assert.Error(t, err, "injectMOD overwrote a function with an empty one")
}

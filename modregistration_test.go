package tela

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestVerifyMODRegistration covers what Verify accepts as a MOD registration.
//
// The count of a MOD's declared function names does not tie those names to its
// code. injectMOD copies sc.Functions[name] for every declared name, so a name
// the code does not define is injected as an empty function instead of being
// rejected at registration.
//
// The MODs values here are local, so this test does not read or write the
// package Mods variable which TestTELA adds to.
func TestVerifyMODRegistration(t *testing.T) {
	class := MODClass{Name: "Test class", Tag: "tc"}
	single := `Function RealName() Uint64
10 RETURN 0
End Function`
	double := `Function RealName() Uint64
10 RETURN 0
End Function

Function OtherName() Uint64
10 RETURN 0
End Function`

	withNames := func(names []string, code string) MODs {
		return MODs{
			mods: []MOD{{
				Name:          "Test mod",
				Tag:           class.NewTag("one"),
				Description:   "registration test",
				FunctionCode:  func() string { return code },
				FunctionNames: names,
			}},
			classes: []MODClass{class},
			index:   []int{1},
		}
	}

	tests := []struct {
		name    string
		mods    MODs
		wantErr bool
	}{
		{"names match the code", withNames([]string{"RealName"}, single), false},
		{"declared name is not in the code", withNames([]string{"NotDefined"}, single), true},
		{"declared names are duplicated", withNames([]string{"RealName", "RealName"}, double), true},
		{"both names match the code", withNames([]string{"RealName", "OtherName"}, double), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mods.Verify()
			if tt.wantErr {
				assert.Error(t, err, "Verify should not accept a MOD where %s", tt.name)
				return
			}

			assert.NoError(t, err, "Verify should accept a MOD where %s: %s", tt.name, err)
		})
	}
}

// TestVerifyMODClassTagOverlap covers class tags which prefix one another.
//
// A MODClass tag prefixes all of its members' tags, and both GetClass and the
// Single MOD rule match on that prefix, so overlapping class tags make two
// classes indistinguishable and cause the Single MOD rule to reject valid tag
// combinations from different classes.
func TestVerifyMODClassTagOverlap(t *testing.T) {
	tests := []struct {
		name    string
		tags    []string
		wantErr bool
	}{
		{"distinct tags", []string{"vs", "tx"}, false},
		{"one tag prefixes another", []string{"vs", "v"}, true},
		{"prefix declared first", []string{"v", "vs"}, true},
		{"shared prefix but neither contains the other", []string{"va", "vb"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var classes []MODClass
			var index []int
			for i, tag := range tt.tags {
				classes = append(classes, MODClass{Name: fmt.Sprintf("Class %d", i), Tag: tag})
				index = append(index, 0)
			}

			m := MODs{classes: classes, index: index}
			err := m.Verify()
			if tt.wantErr {
				assert.Error(t, err, "Verify should not accept class tags %v", tt.tags)
				return
			}

			assert.NoError(t, err, "Verify should accept class tags %v: %s", tt.tags, err)
		})
	}
}

// TestMODsRegisteredAtInit covers what initMods left registered.
//
// Add adds nothing when any part of what it is given conflicts, so one bad
// entry drops a whole MODClass and every MOD in it, and the Verify below passes
// because what remains is consistent. Both Add calls report that now, but the
// registration itself is built from literals with no seam a test can fail, so
// this pins the result rather than the report.
//
// Only presence is asserted, which is what a silent drop takes away. Anything
// registered afterwards is additive and does not affect it.
func TestMODsRegisteredAtInit(t *testing.T) {
	for _, tag := range []string{"vs", "tx"} {
		var found bool
		for _, c := range Mods.GetAllClasses() {
			if c.Tag == tag {
				found = true
				break
			}
		}

		assert.True(t, found, "MODClass %q was not registered", tag)
	}

	for _, tag := range []string{"vsoo", "vsooim", "vspubsu", "vspubow", "vspubim", "txdwd", "txdwa", "txto"} {
		assert.NotEmpty(t, Mods.GetMod(tag).Name, "MOD %q was not registered", tag)
	}
}

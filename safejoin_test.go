package tela

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSafeJoin covers the containment guard on the clone/serve path. Every input
// here is a value a smart contract controls - a dURL, a subDir, or a header file
// name - and each reaches an os.Create/os.WriteFile.
func TestSafeJoin(t *testing.T) {
	base := filepath.Join("/tmp", "datashards", "clone")

	t.Run("Escapes are rejected", func(t *testing.T) {
		bad := []string{
			"../../../../../../tmp/pwned", // the reproduced dURL exploit
			"../evil.desktop",             // one level up, parent exists
			"..",
			"a/../../b",
			"sub/../../../etc/x",
		}
		for _, e := range bad {
			_, err := safeJoin(base, e)
			assert.Error(t, err, "must reject %q", e)
		}
	})

	t.Run("Legitimate paths are allowed", func(t *testing.T) {
		good := []string{
			"villager.tela",
			"villager/rive.js-2.35.3.shards", // a real mainnet dURL with a separator
			"index.html",
			"sub1/sub2/file.js",
			"a/../b", // stays within base
			"",
			".",
		}
		for _, e := range good {
			got, err := safeJoin(base, e)
			assert.NoError(t, err, "must allow %q", e)
			// and the result must actually be within base
			assert.True(t, got == base || strings.HasPrefix(got, base+string(filepath.Separator)),
				"%q resolved outside base: %s", e, got)
		}
	})

	t.Run("Absolute element is contained, not honored", func(t *testing.T) {
		got, err := safeJoin(base, "/etc/passwd")
		assert.NoError(t, err)
		assert.True(t, strings.HasPrefix(got, base+string(filepath.Separator)), "absolute path escaped: %s", got)
	})
}

// TestConstructFromShardsContainment pins the specific site O19 reopened on:
// the .shards clone path, which the first containment sweep missed.
func TestConstructFromShardsContainment(t *testing.T) {
	base := t.TempDir()
	err := ConstructFromShards([][]byte{[]byte("payload")}, "../escaped.desktop", base, "")
	assert.Error(t, err, "ConstructFromShards must reject a traversal file name")
	assert.NoFileExists(t, filepath.Join(filepath.Dir(base), "escaped.desktop"))
}

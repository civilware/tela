package tela

import (
	"encoding/hex"
	"testing"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/transaction"
	"github.com/stretchr/testify/assert"
)

// scidHex returns a deterministic 64 character SCID for tests
func scidHex(b byte) string {
	var h crypto.Hash
	for i := range h {
		h[i] = b
	}

	return hex.EncodeToString(h[:])
}

// scTX builds a SC transaction carrying args
func scTX(args rpc.Arguments) *transaction.Transaction {
	var tx transaction.Transaction
	tx.Version = 1
	tx.TransactionType = transaction.SC_TX
	tx.SCDATA = args

	return &tx
}

// TestCommitSCID covers which contract a checkout TXID actually applies to.
//
// cloneINDEXAtCommit reads the dURL from the SCID but rebuilds the code from the
// TXID. Without comparing the two, asking for one contract's history with
// another contract's TXID serves the second contract's files under the first
// contract's dURL.
func TestCommitSCID(t *testing.T) {
	scidA := scidHex(0xaa)
	scidB := scidHex(0xbb)

	var targetB crypto.Hash
	raw, err := hex.DecodeString(scidB)
	assert.NoError(t, err, "Test SCID should decode")
	copy(targetB[:], raw)

	t.Run("Install", func(t *testing.T) {
		// A contract's SCID is the hash of the transaction that installed it
		tx := scTX(rpc.Arguments{
			{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_INSTALL)},
			{Name: rpc.SCCODE, DataType: rpc.DataString, Value: "Function Initialize() Uint64\n10 RETURN 0\nEnd Function"},
		})

		scid, err := commitSCID(tx, scidA)
		assert.NoError(t, err, "Install should be a valid commit")
		assert.Equal(t, scidA, scid, "Install commit should apply to its own TXID")
	})

	t.Run("Call", func(t *testing.T) {
		tx := scTX(rpc.Arguments{
			{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
			{Name: rpc.SCID, DataType: rpc.DataHash, Value: targetB},
			{Name: "entrypoint", DataType: rpc.DataString, Value: "UpdateCode"},
		})

		// Called with A's TXID, but the transaction names B
		scid, err := commitSCID(tx, scidA)
		assert.NoError(t, err, "Call should be a valid commit")
		assert.Equal(t, scidB, scid, "Call commit should apply to the SCID it names, not the TXID")
	})

	t.Run("Not a commit", func(t *testing.T) {
		var normal transaction.Transaction
		normal.Version = 1
		normal.TransactionType = transaction.NORMAL
		_, err := commitSCID(&normal, scidA)
		assert.Error(t, err, "Non SC transaction should not be a commit")

		_, err = commitSCID(scTX(rpc.Arguments{
			{Name: "entrypoint", DataType: rpc.DataString, Value: "UpdateCode"},
		}), scidA)
		assert.Error(t, err, "SC transaction with no action should not be a commit")

		_, err = commitSCID(scTX(rpc.Arguments{
			{Name: rpc.SCACTION, DataType: rpc.DataUint64, Value: uint64(rpc.SC_CALL)},
		}), scidA)
		assert.Error(t, err, "Call naming no contract should not be a commit")
	})
}

// TestCommitMatchesSCID covers the binding cloneINDEXAtCommit applies before it
// rebuilds a contract's code from a TXID.
func TestCommitMatchesSCID(t *testing.T) {
	scidA := scidHex(0xaa)

	t.Run("Malformed", func(t *testing.T) {
		err := commitMatchesSCID("nothex", scidA, scidA)
		assert.Error(t, err, "Undecodable TXID data should be rejected")

		err = commitMatchesSCID("aabbcc", scidA, scidA)
		assert.Error(t, err, "Undeserializable TXID data should be rejected")

		err = commitMatchesSCID("", scidA, scidA)
		assert.Error(t, err, "Empty TXID data should be rejected")
	})
}

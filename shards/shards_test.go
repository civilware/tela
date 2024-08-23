package shards

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/walletapi"
	"github.com/deroproject/graviton"
	"github.com/stretchr/testify/assert"
)

func TestShards(t *testing.T) {
	mainPath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error: %s", err)
	}

	endpoint := "127.0.0.1:20000"
	walletName := "tela-store"
	testPath := filepath.Join(mainPath, "tela_tests")
	datashards := filepath.Join(testPath, "datashards")
	walletPath := filepath.Join(testPath, "tela_sim_wallets")
	walletSeeds := []string{
		"2af6630905d73ee40864bd48339f297908a0731a6c4c6fa0a27ea574ac4e4733",
		"193faf64d79e9feca5fce8b992b4bb59b86c50f491e2dc475522764ca6666b6b",
	}

	t.Cleanup(func() {
		os.RemoveAll(testPath)
	})

	os.RemoveAll(testPath)

	// Create simulator wallets to use for encrypt/decrypt tests
	var wallets []*walletapi.Wallet_Disk
	for i, seed := range walletSeeds {
		w, err := createTestWallet(fmt.Sprintf("%s%d", walletName, i), walletPath, seed)
		if err != nil {
			t.Fatalf("Failed to create wallet %d: %s", i, err)
		}

		wallets = append(wallets, w)
		assert.NotNil(t, wallets[i], "Wallet %s should not be nil", i)
	}

	t.Run("Store", func(t *testing.T) {
		// Test SetPath
		testPath := "likely/A/invalid/Path"
		_, err = SetPath(testPath)
		assert.Error(t, err, "Invalid path should not be set")
		defaultPath := filepath.Join(mainPath, "datashards")
		assert.Equal(t, defaultPath, GetPath(), "Datashard path should have remained as default")
		_, err = SetPath(datashards) // set test path
		assert.NoError(t, err, "Valid path should not error: %s", err)

		// Values to store
		tree := "user"
		key := []byte("other_network")
		value := []byte(endpoint + endpoint)

		for _, dbType := range dbTypes {
			err := SetDBType(dbType)
			assert.NoError(t, err, "Setting db type should not error: %s", err)

			t.Run("StoreValue", func(t *testing.T) {
				err := StoreValue(tree, key, value)
				assert.NoError(t, err, "Storing value should not error: %s", err)
			})

			t.Run("GetValue", func(t *testing.T) {
				result, err := GetValue(tree, key)
				assert.NoError(t, err, "Getting stored value should not error: %s", err)
				assert.Equal(t, value, result, "Stored endpoints should be equal")
			})

			t.Run("SettingsValue", func(t *testing.T) {
				err := StoreSettingsValue(key, value)
				assert.NoError(t, err, "Storing value should not error: %s", err)
				result, err := GetSettingsValue(key)
				assert.NoError(t, err, "Getting stored value should not error: %s", err)
				assert.Equal(t, value, result, "Stored endpoints should be equal")
			})

			t.Run("Endpoint", func(t *testing.T) {
				endpoint := "endpoint"
				err := StoreEndpoint(endpoint)
				assert.NoError(t, err, "Storing endpoint value should not error: %s", err)
				result, err := GetEndpoint()
				assert.NoError(t, err, "Getting endpoint should not error: %s", err)
				assert.Equal(t, endpoint, result, "Stored endpoints should be equal")
			})

			t.Run("Network", func(t *testing.T) {
				network := "Testnet"
				err := StoreNetwork(network)
				assert.NoError(t, err, "Storing network value should not error: %s", err)
				result, err := GetNetwork()
				assert.NoError(t, err, "Getting network should not error: %s", err)
				assert.Equal(t, network, result, "Stored networks should be equal")
			})

			t.Run("StoreEncryptedValue", func(t *testing.T) {
				err := StoreEncryptedValue(wallets[0], tree, key, value)
				assert.NoError(t, err, "Storing encrypted value should not error: %s", err)
			})

			t.Run("GetEncryptedValue", func(t *testing.T) {
				result, err := GetEncryptedValue(wallets[0], tree, key)
				assert.NoError(t, err, "Getting encrypted store value should not error: %s", err)
				assert.Equal(t, value, result, "Stored decrypted values should be equal")
			})

			t.Run("DeleteKey", func(t *testing.T) {
				err = DeleteKey(wallets[0], tree, key)
				assert.NoError(t, err, "Deleting valid key should not error: %s", err)
			})

			t.Run("DeleteSettingsKey", func(t *testing.T) {
				err = DeleteSettingsKey(key)
				assert.NoError(t, err, "Deleting valid settings key should not error: %s", err)
			})
		}

		for _, db := range dbTypes {
			assert.True(t, IsValidDBType(db), "%s should be a valid DB type", db)
		}

		// Test invalid DB types
		invalidDB := "invalidDB"
		assert.False(t, IsValidDBType(invalidDB), "%s should not be a valid DB type", invalidDB)

		err = SetDBType(invalidDB)
		assert.Error(t, err, "Setting invalid db type should error")
		assert.Equal(t, "boltdb", GetDBType(), "DB type should be boltdb: %s", GetDBType())
		// SetDBType should not allow this to occur
		shards.dbType = "unlikely"
		err := StoreValue(tree, key, value)
		assert.Error(t, err, "Storing value with invalid DB type should error")
		_, err = GetValue(tree, key)
		assert.Error(t, err, "Getting value with invalid DB type should error")
		err = StoreSettingsValue(key, value)
		assert.Error(t, err, "Storing settings value with invalid DB type should error")
		_, err = GetSettingsValue(key)
		assert.Error(t, err, "Getting settings value with invalid DB type should error")
		err = StoreEndpoint("endpoint")
		assert.Error(t, err, "Storing endpoint with invalid DB type should error")
		_, err = GetEndpoint()
		assert.Error(t, err, "Getting endpoint with invalid DB type should error")
		err = StoreNetwork("Mainnet")
		assert.Error(t, err, "Storing valid network with invalid DB type should error")
		err = StoreNetwork("invalid")
		assert.Error(t, err, "Storing invalid network with invalid DB type should error")
		_, err = GetNetwork()
		assert.Error(t, err, "Getting network with invalid DB type should error")
		err = StoreEncryptedValue(wallets[0], tree, key, value)
		assert.Error(t, err, "Storing encrypted value with invalid DB type should error")
		_, err = GetEncryptedValue(wallets[0], tree, key)
		assert.Error(t, err, "Getting encrypted value with invalid DB type should error")
		err = DeleteKey(wallets[0], tree, key)
		assert.Error(t, err, "Deleting key value with invalid DB type should error")
		err = DeleteSettingsKey(key)
		assert.Error(t, err, "Deleting settings key value with invalid DB type should error")
	})

	// // Test bbolt functions
	t.Run("Bolt", func(t *testing.T) {
		// Values to store
		bucket := "settings"
		key := []byte("network")
		value := []byte(endpoint)

		// Test boltStoreValue
		t.Run("boltStoreValue", func(t *testing.T) {
			err := boltStoreValue(bucket, key, value)
			assert.NoError(t, err, "Storing value should not error: %s", err)
			// boltStoreValue errors
			err = boltStoreValue("", key, value)
			assert.Error(t, err, "Storing empty bucket should error")
			err = boltStoreValue(bucket, nil, value)
			assert.Error(t, err, "Storing invalid key should error")
		})

		// Test boltGetValue
		t.Run("boltGetValue", func(t *testing.T) {
			result, err := boltGetValue(bucket, key)
			assert.NoError(t, err, "Getting stored value should not error: %s", err)
			assert.Equal(t, value, result, "Stored endpoints should be equal")
			// boltGetValue errors
			_, err = boltGetValue("", key)
			assert.Error(t, err, "Getting empty bucket should error")
			_, err = boltGetValue(bucket, nil)
			assert.Error(t, err, "Getting empty value should error")
		})

		// Test boltStoreEncryptedValue
		t.Run("boltStoreEncryptedValue", func(t *testing.T) {
			err := boltStoreEncryptedValue(wallets[0], bucket, key, value)
			assert.NoError(t, err, "Storing encrypted value should not error: %s", err)
			// boltStoreEncryptedValue errors
			err = boltStoreEncryptedValue(wallets[0], "", key, value)
			assert.Error(t, err, "Storing encrypted value with empty bucket should error")
			err = boltStoreEncryptedValue(wallets[0], bucket, nil, value)
			assert.Error(t, err, "Storing encrypted value with nil key should error")
		})

		// Test boltGetEncryptedValue
		t.Run("boltGetEncryptedValue", func(t *testing.T) {
			result, err := boltGetEncryptedValue(wallets[0], bucket, key)
			assert.NoError(t, err, "Getting encrypted store value should not error: %s", err)
			assert.Equal(t, value, result, "Stored decrypted values should be equal")
			// boltGetEncryptedValue errors
			_, err = boltGetEncryptedValue(wallets[0], "", key)
			assert.Error(t, err, "Getting encrypted value with empty bucket should error")
			_, err = boltGetEncryptedValue(wallets[0], bucket, nil)
			assert.Error(t, err, "Getting encrypted value with nil key should error")
		})

		// Test boltDeleteKey
		t.Run("boltDeleteKey", func(t *testing.T) {
			err = boltDeleteKey(nil, bucket, key)
			assert.NoError(t, err, "Deleting settings with valid key should not error: %s", err)
			err = boltDeleteKey(nil, "", key)
			assert.Error(t, err, "Deleting settings with empty bucket error")
			// boltDeleteKey errors
			err = boltDeleteKey(wallets[0], "", key)
			assert.Error(t, err, "Deleting shard with empty bucket should error")
			err = boltDeleteKey(wallets[0], bucket, key)
			assert.NoError(t, err, "Deleting shard with valid key should not error: %s", err)
			err = boltDeleteKey(wallets[1], bucket, key)
			assert.Error(t, err, "Deleting non existent bucket should error")
		})
	})

	// // Test Graviton functions
	t.Run("Graviton", func(t *testing.T) {
		// Values to store
		tree := "settings"
		key := []byte("network")
		value := []byte(endpoint)

		// For errors
		treeToBig := strings.Repeat("t", graviton.TREE_NAME_LIMIT+1)
		valueToBIg := []byte(strings.Repeat("t", graviton.MAX_VALUE_SIZE+1))

		// Test gravitonStoreValue
		t.Run("gravitonStoreValue", func(t *testing.T) {
			err := gravitonStoreValue(tree, key, value)
			assert.NoError(t, err, "Storing settings value should not error: %s", err)
			// gravitonStoreValue errors
			err = gravitonStoreValue("", key, value)
			assert.Error(t, err, "Storing empty tree should error")
			err = gravitonStoreValue(tree, nil, nil)
			assert.Error(t, err, "Storing nil key should error")
			err = gravitonStoreValue(tree, key, nil)
			assert.Error(t, err, "Storing nil value should error")
			err = gravitonStoreValue(treeToBig, key, value)
			assert.Error(t, err, "Storing value with tree exceeding %d should error", graviton.TREE_NAME_LIMIT)
			err = gravitonStoreValue(tree, key, valueToBIg)
			assert.Error(t, err, "Storing value with value exceeding %d should error", graviton.MAX_VALUE_SIZE)
		})

		// Test gravitonGetValue
		t.Run("gravitonGetValue", func(t *testing.T) {
			result, err := gravitonGetValue(tree, key)
			assert.NoError(t, err, "Getting settings value should not error: %s", err)
			assert.Equal(t, value, result, "Stored endpoints should be equal")
			// gravitonGetValue errors
			_, err = gravitonGetValue("", key)
			assert.Error(t, err, "Getting empty tree should error")
			_, err = gravitonGetValue(tree, nil)
			assert.Error(t, err, "Getting nil key should error")
			_, err = gravitonGetValue(treeToBig, key)
			assert.Error(t, err, "Getting value with tree to big should error")
			_, err = gravitonGetValue(tree, []byte("not here"))
			assert.Error(t, err, "Getting value non existent key should error")
		})

		// Test gravitonStoreEncryptedValue
		t.Run("gravitonStoreEncryptedValue", func(t *testing.T) {
			err := gravitonStoreEncryptedValue(wallets[0], tree, key, value)
			assert.NoError(t, err, "Storing encrypted value should not error: %s", err)
			// gravitonStoreEncryptedValue errors
			err = gravitonStoreEncryptedValue(wallets[0], "", key, value)
			assert.Error(t, err, "Storing encrypted value with empty tree should error")
			err = gravitonStoreEncryptedValue(wallets[0], tree, nil, nil)
			assert.Error(t, err, "Storing encrypted value with nil key should error")
			err = gravitonStoreEncryptedValue(wallets[0], tree, key, nil)
			assert.Error(t, err, "Storing encrypted value with nil value should error")
			err = gravitonStoreEncryptedValue(nil, tree, key, nil)
			assert.Error(t, err, "Storing encrypted value should error with nil wallet")
			err = gravitonStoreEncryptedValue(wallets[0], treeToBig, key, value)
			assert.Error(t, err, "Storing encrypted value with tree exceeding %d should error", graviton.TREE_NAME_LIMIT)
			err = gravitonStoreEncryptedValue(wallets[0], tree, key, valueToBIg)
			assert.Error(t, err, "Storing encrypted value with value exceeding %d should error", graviton.MAX_VALUE_SIZE)
		})

		// Test gravitonGetEncryptedValue
		t.Run("gravitonGetEncryptedValue", func(t *testing.T) {
			result, err := gravitonGetEncryptedValue(wallets[0], tree, key)
			assert.NoError(t, err, "Getting encrypted store value should not error: %s", err)
			assert.Equal(t, value, result, "Stored decrypted values should be equal")
			// gravitonGetEncryptedValue errors
			_, err = gravitonGetEncryptedValue(wallets[0], "", key)
			assert.Error(t, err, "Getting empty tree should error")
			_, err = gravitonGetEncryptedValue(wallets[0], treeToBig, key)
			assert.Error(t, err, "Getting encrypted value with tree to big should error")
			_, err = gravitonGetEncryptedValue(wallets[0], tree, nil)
			assert.Error(t, err, "Getting nil key should error")
			_, err = gravitonGetEncryptedValue(nil, tree, key)
			assert.Error(t, err, "Getting encrypted value should error with nil wallet")
			_, err = gravitonGetEncryptedValue(wallets[1], tree, key)
			assert.Error(t, err, "Encrypted value should not exists for this wallet")
		})

		// Test gravitonDeleteKey
		t.Run("gravitonDeleteKey", func(t *testing.T) {
			err := gravitonDeleteKey(wallets[0], tree, key)
			assert.NoError(t, err, "Deleting valid key should not error: %s", err)
			// gravitonDeleteKey errors
			err = gravitonDeleteKey(wallets[0], "", key)
			assert.Error(t, err, "Deleting with empty tree should error")
			err = gravitonDeleteKey(wallets[0], tree, nil)
			assert.Error(t, err, "Deleting with nil key should error")
			err = gravitonDeleteKey(nil, tree, nil)
			assert.Error(t, err, "Deleting with nil wallet should error")
			err = gravitonDeleteKey(wallets[0], treeToBig, key)
			assert.Error(t, err, "Deleting with with value exceeding %d should error", graviton.TREE_NAME_LIMIT)
		})
	})

	t.Run("Keys", func(t *testing.T) {
		assert.Equal(t, []byte("endpoint"), Key.Endpoint(), "Endpoint keys should be equal")
		assert.Equal(t, []byte("network"), Key.Network(), "Network keys should be equal")
	})

	t.Run("Values", func(t *testing.T) {
		assert.Equal(t, "Mainnet", Value.Network.Mainnet(), "Mainnet values should be equal")
		assert.Equal(t, "Testnet", Value.Network.Testnet(), "Testnet values should be equal")
		assert.Equal(t, "Simulator", Value.Network.Simulator(), "Simulator values should be equal")
	})
}

// Create test wallet
func createTestWallet(name, dir, seed string) (wallet *walletapi.Wallet_Disk, err error) {
	seed_raw, err := hex.DecodeString(seed)
	if err != nil {
		return
	}

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return
	}

	filename := filepath.Join(dir, name)

	os.Remove(filename)

	wallet, err = walletapi.Create_Encrypted_Wallet(filename, "", new(crypto.BNRed).SetBytes(seed_raw))
	if err != nil {
		return
	}

	wallet.SetNetwork(false)
	wallet.Save_Wallet()

	return
}

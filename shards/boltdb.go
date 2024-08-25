package shards

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/deroproject/derohe/walletapi"
	"go.etcd.io/bbolt"
)

// Get shard and make directory for DB
func boltGetShard(disk *walletapi.Wallet_Disk) (shard string, err error) {
	shard = GetShard(disk)
	err = os.MkdirAll(shard, os.ModePerm)
	return
}

// Encrypt a key-value and then store it in a bbolt DB
func boltStoreEncryptedValue(disk *walletapi.Wallet_Disk, bucket string, key []byte, value []byte) (err error) {
	var shard string
	shard, err = boltGetShard(disk)
	if err != nil {
		return
	}

	var db *bbolt.DB
	db, err = bbolt.Open(filepath.Join(shard, filepath.Base(shard)+".db"), 0600, nil)
	if err != nil {
		return
	}

	err = db.Update(func(tx *bbolt.Tx) (err error) {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return
		}

		eValue, err := disk.Encrypt(value)
		if err != nil {
			return
		}

		return b.Put(key, eValue)
	})

	db.Close()

	return
}

// Store a key-value in a bbolt DB
func boltStoreValue(bucket string, key []byte, value []byte) (err error) {
	var shard string
	shard, err = boltGetShard(nil)
	if err != nil {
		return
	}

	var db *bbolt.DB
	db, err = bbolt.Open(filepath.Join(shard, "settings.db"), 0600, nil)
	if err != nil {
		return
	}

	err = db.Update(func(tx *bbolt.Tx) (err error) {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return
		}

		mar, err := json.Marshal(&value)
		if err != nil {
			return
		}

		return b.Put(key, mar)
	})

	db.Close()

	return
}

// Get a key-value from a bbolt DB
func boltGetValue(bucket string, key []byte) (result []byte, err error) {
	var shard string
	shard, err = boltGetShard(nil)
	if err != nil {
		return
	}

	var db *bbolt.DB
	db, err = bbolt.Open(filepath.Join(shard, "settings.db"), 0600, nil)
	if err != nil {
		return
	}

	err = db.View(func(tx *bbolt.Tx) (err error) {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}

		stored := b.Get(key)
		if stored != nil {
			return json.Unmarshal(stored, &result)
		}

		return fmt.Errorf("key %q not found", key)
	})

	db.Close()

	return
}

// Store a key-value in bbolt DB settings bucket
func boltStoreSettingsValue(key, value []byte) (err error) {
	return boltStoreValue("settings", key, value)
}

// Get a key-value from bbolt DB settings bucket
func boltGetSettingsValue(key []byte) (result []byte, err error) {
	return boltGetValue("settings", key)
}

// Store endpoint value in bbolt DB settings bucket
func boltStoreEndpoint(value []byte) (err error) {
	return boltStoreSettingsValue(Key.Endpoint(), value)
}

// Get endpoint value from bbolt DB settings bucket
func boltGetEndpoint() (result []byte, err error) {
	return boltGetSettingsValue(Key.Endpoint())
}

// Store network value in bbolt DB settings bucket
func boltStoreNetwork(value []byte) (err error) {
	return boltStoreSettingsValue(Key.Network(), value)
}

// Get network value from bbolt DB settings bucket
func boltGetNetwork() (result []byte, err error) {
	return boltGetSettingsValue(Key.Network())
}

// Get an encrypted key-value from a bbolt DB and then decrypt it
func boltGetEncryptedValue(disk *walletapi.Wallet_Disk, bucket string, key []byte) (result []byte, err error) {
	var shard string
	shard, err = boltGetShard(disk)
	if err != nil {
		return
	}

	var db *bbolt.DB
	db, err = bbolt.Open(filepath.Join(shard, filepath.Base(shard)+".db"), 0600, nil)
	if err != nil {
		return
	}

	err = db.View(func(tx *bbolt.Tx) (err error) {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}

		stored := b.Get(key)
		if stored == nil {
			return fmt.Errorf("key %q not found", key)
		}

		eValue, err := disk.Decrypt(stored)
		if err != nil {
			return
		}

		result = eValue

		return
	})

	db.Close()

	return
}

// Delete a key-value in a bbolt DB
func boltDeleteKey(disk *walletapi.Wallet_Disk, bucket string, key []byte) (err error) {
	shard := GetShard(disk)

	var db *bbolt.DB
	db, err = bbolt.Open(filepath.Join(shard, filepath.Base(shard)+".db"), 0600, nil)
	if err != nil {
		return
	}

	err = db.Update(func(tx *bbolt.Tx) (err error) {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}

		return b.Delete(key)
	})

	db.Close()

	return
}

package shards

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/deroproject/derohe/walletapi"
)

type dbstore struct {
	path   string
	dbType string
	sync.RWMutex
}

var shards dbstore
var dbTypes = []string{"gravdb", "boltdb"}

// Initialize package defaults and datashard storage location
func init() {
	shards.dbType = "gravdb"
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("[shards] %s\n", err)
		// Fallback location
		shards.path = "datashards"
	} else {
		shards.path = filepath.Join(dir, "datashards")
	}
}

// SetPath can be used to set a custom path for datashard storage
func SetPath(path string) (datashards string, err error) {
	dir := filepath.Dir(path)
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = fmt.Errorf("%s does not exists", path)
		return
	}

	datashards = filepath.Join(path, "datashards")
	shards.Lock()
	shards.path = datashards
	shards.Unlock()

	return
}

// Get the current datashards storage path
func GetPath() string {
	shards.RLock()
	defer shards.RUnlock()

	return shards.path
}

// Is db a valid type for datashard storage
func IsValidDBType(db string) bool {
	shards.RLock()
	defer shards.RUnlock()

	for _, t := range dbTypes {
		if db == t {
			return true
		}
	}

	return false
}

// Set datashards DB type if IsValidDBType
func SetDBType(db string) (err error) {
	shards.Lock()
	defer shards.Unlock()

	switch db {
	case "gravdb":
		shards.dbType = db
	case "boltdb":
		shards.dbType = db
	default:
		return fmt.Errorf("unknown db type %q", db)
	}

	return
}

// Get datashards DB type
func GetDBType() string {
	shards.RLock()
	defer shards.RUnlock()

	return shards.dbType
}

// Get a datashard's path, synced with Engram's DB structure
func GetShard(disk *walletapi.Wallet_Disk) (result string) {
	dir := GetPath()
	if disk == nil {
		result = filepath.Join(dir, "settings")
	} else {
		address := disk.GetAddress().String()
		result = filepath.Join(dir, fmt.Sprintf("%x", sha1.Sum([]byte(address))))
	}

	return
}

// Encrypt a key-value and then store it in datashards
func StoreEncryptedValue(disk *walletapi.Wallet_Disk, t string, key, value []byte) (err error) {
	db := GetDBType()
	switch db {
	case "gravdb":
		err = gravitonStoreEncryptedValue(disk, t, key, value)
	case "boltdb":
		err = boltStoreEncryptedValue(disk, t, key, value)
	default:
		err = fmt.Errorf("unknown db type %q", db)
	}

	return
}

// Store a key-value in datashards
func StoreValue(t string, key, value []byte) (err error) {
	db := GetDBType()
	switch db {
	case "gravdb":
		err = gravitonStoreValue(t, key, value)
	case "boltdb":
		err = boltStoreValue(t, key, value)
	default:
		err = fmt.Errorf("unknown db type %q", db)
	}

	return
}

// Get a key-value from datashards
func GetValue(t string, key []byte) (result []byte, err error) {
	db := GetDBType()
	switch db {
	case "gravdb":
		result, err = gravitonGetValue(t, key)
	case "boltdb":
		result, err = boltGetValue(t, key)
	default:
		err = fmt.Errorf("unknown db type %q", db)
	}

	return
}

// Store a key-value setting in datashards
func StoreSettingsValue(key, value []byte) (err error) {
	db := GetDBType()
	switch db {
	case "gravdb":
		err = gravitonStoreSettingsValue(key, value)
	case "boltdb":
		err = boltStoreSettingsValue(key, value)
	default:
		err = fmt.Errorf("unknown db type %q", db)
	}

	return
}

// Get a key-value setting from datashards
func GetSettingsValue(key []byte) (result []byte, err error) {
	db := GetDBType()
	switch db {
	case "gravdb":
		result, err = gravitonGetSettingsValue(key)
	case "boltdb":
		result, err = boltGetSettingsValue(key)
	default:
		err = fmt.Errorf("unknown db type %q", db)
	}

	return
}

// Store endpoint value in datashards
func StoreEndpoint(value string) (err error) {
	db := GetDBType()
	switch db {
	case "gravdb":
		err = gravitonStoreEndpoint([]byte(value))
	case "boltdb":
		err = boltStoreEndpoint([]byte(value))
	default:
		err = fmt.Errorf("unknown db type %q", db)
	}

	return
}

// Get endpoint value from datashards
func GetEndpoint() (endpoint string, err error) {
	var result []byte
	db := GetDBType()
	switch db {
	case "gravdb":
		result, err = gravitonGetEndpoint()
	case "boltdb":
		result, err = boltGetEndpoint()
	default:
		err = fmt.Errorf("unknown db type %q", db)
		return
	}

	endpoint = string(result)

	return
}

// Store network value in datashards
func StoreNetwork(value string) (err error) {
	if !isNetworkValid(value) {
		err = fmt.Errorf("invalid network")
		return
	}

	db := GetDBType()
	switch db {
	case "gravdb":
		err = gravitonStoreNetwork([]byte(value))
	case "boltdb":
		err = boltStoreNetwork([]byte(value))
	default:
		err = fmt.Errorf("unknown db type %q", db)
	}

	return
}

// Get network value from datashards
func GetNetwork() (network string, err error) {
	var result []byte
	db := GetDBType()
	switch db {
	case "gravdb":
		result, err = gravitonGetNetwork()
	case "boltdb":
		result, err = boltGetNetwork()
	default:
		err = fmt.Errorf("unknown db type %q", db)
		return
	}

	network = string(result)

	return
}

// Get an encrypted key-value from datashards and then decrypt it
func GetEncryptedValue(disk *walletapi.Wallet_Disk, t string, key []byte) (result []byte, err error) {
	db := GetDBType()
	switch db {
	case "gravdb":
		result, err = gravitonGetEncryptedValue(disk, t, key)
	case "boltdb":
		result, err = boltGetEncryptedValue(disk, t, key)
	default:
		err = fmt.Errorf("unknown db type %q", db)
	}

	return
}

// Delete a key-value from datashards
func DeleteKey(disk *walletapi.Wallet_Disk, t string, key []byte) (err error) {
	db := GetDBType()
	switch db {
	case "gravdb":
		err = gravitonDeleteKey(disk, t, key)
	case "boltdb":
		err = boltDeleteKey(disk, t, key)
	default:
		err = fmt.Errorf("unknown db type %q", db)
	}

	return
}

// Delete a key-value setting from datashards
func DeleteSettingsKey(key []byte) (err error) {
	db := GetDBType()
	switch db {
	case "gravdb":
		err = gravitonDeleteKey(nil, "settings", key)
	case "boltdb":
		err = boltDeleteKey(nil, "settings", key)
	default:
		err = fmt.Errorf("unknown db type %q", db)
	}

	return
}

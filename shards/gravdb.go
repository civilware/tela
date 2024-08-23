package shards

import (
	"fmt"

	"github.com/deroproject/derohe/walletapi"
	"github.com/deroproject/graviton"
)

// Encrypt a key-value and then store it in a Graviton tree
func gravitonStoreEncryptedValue(disk *walletapi.Wallet_Disk, t string, key, value []byte) (err error) {
	if disk == nil {
		err = fmt.Errorf("no active account found")
		return
	}

	if t == "" {
		err = fmt.Errorf("missing graviton tree input")
		return
	}

	if key == nil {
		err = fmt.Errorf("missing graviton key input")
		return
	}

	if value == nil {
		err = fmt.Errorf("missing graviton value input")
		return
	}

	eValue, err := disk.Encrypt(value)
	if err != nil {
		return
	}

	store, err := graviton.NewDiskStore(GetShard(disk))
	if err != nil {
		return
	}

	ss, err := store.LoadSnapshot(0)
	if err != nil {
		return
	}

	tree, err := ss.GetTree(t)
	if err != nil {
		return
	}

	err = tree.Put(key, eValue)
	if err != nil {
		return
	}

	_, err = graviton.Commit(tree)
	if err != nil {
		return
	}

	return
}

// Store a key-value in a Graviton tree
func gravitonStoreValue(t string, key, value []byte) (err error) {
	if t == "" {
		err = fmt.Errorf("missing graviton tree input")
		return
	}

	if key == nil {
		err = fmt.Errorf("missing graviton key input")
		return
	}

	if value == nil {
		err = fmt.Errorf("missing graviton value input")
		return
	}

	store, err := graviton.NewDiskStore(GetShard(nil))
	if err != nil {
		return
	}

	ss, err := store.LoadSnapshot(0)
	if err != nil {
		return
	}

	tree, err := ss.GetTree(t)
	if err != nil {
		return
	}

	err = tree.Put(key, value)
	if err != nil {
		return
	}

	_, err = graviton.Commit(tree)
	if err != nil {
		return
	}

	return
}

// Get a key-value from a Graviton tree
func gravitonGetValue(t string, key []byte) (result []byte, err error) {
	result = []byte("")
	if t == "" {
		err = fmt.Errorf("missing graviton tree input")
		return
	}

	if key == nil {
		err = fmt.Errorf("missing graviton key input")
		return
	}

	store, err := graviton.NewDiskStore(GetShard(nil))
	if err != nil {
		return
	}

	ss, err := store.LoadSnapshot(0)
	if err != nil {
		return
	}

	tree, err := ss.GetTree(t)
	if err != nil {
		return
	}

	result, err = tree.Get(key)
	if err != nil {
		return
	}

	return
}

// Store a key-value in Graviton settings tree
func gravitonStoreSettingsValue(key, value []byte) (err error) {
	return gravitonStoreValue("settings", key, value)
}

// Get a key-value from settings Graviton tree
func gravitonGetSettingsValue(key []byte) (result []byte, err error) {
	return gravitonGetValue("settings", key)
}

// Store endpoint value in Graviton settings tree
func gravitonStoreEndpoint(value []byte) (err error) {
	return gravitonStoreSettingsValue(Key.Endpoint(), value)
}

// Get endpoint value from settings Graviton tree
func gravitonGetEndpoint() (result []byte, err error) {
	return gravitonGetSettingsValue(Key.Endpoint())
}

// Store network value in Graviton settings tree
func gravitonStoreNetwork(value []byte) (err error) {
	return gravitonStoreSettingsValue(Key.Network(), value)
}

// Get network value from settings Graviton tree
func gravitonGetNetwork() (result []byte, err error) {
	return gravitonGetSettingsValue(Key.Network())
}

// Get an encrypted key-value from a Graviton tree and then decrypt it
func gravitonGetEncryptedValue(disk *walletapi.Wallet_Disk, t string, key []byte) (result []byte, err error) {
	result = []byte("")
	if disk == nil {
		err = fmt.Errorf("no active account found")
		return
	}

	if t == "" {
		err = fmt.Errorf("missing graviton tree input")
		return
	}

	if key == nil {
		err = fmt.Errorf("missing graviton key input")
		return
	}

	store, err := graviton.NewDiskStore(GetShard(disk))
	if err != nil {
		return
	}

	ss, err := store.LoadSnapshot(0)
	if err != nil {
		return
	}

	tree, err := ss.GetTree(t)
	if err != nil {
		return
	}

	eValue, err := tree.Get(key)
	if err != nil {
		return
	}

	result, err = disk.Decrypt(eValue)
	if err != nil {
		return
	}

	return
}

// Delete a key-value in a Graviton tree
func gravitonDeleteKey(disk *walletapi.Wallet_Disk, t string, key []byte) (err error) {
	if t == "" {
		err = fmt.Errorf("missing graviton tree input")
		return
	}

	if key == nil {
		err = fmt.Errorf("missing graviton key input")
		return
	}

	store, err := graviton.NewDiskStore(GetShard(disk))
	if err != nil {
		return
	}

	ss, err := store.LoadSnapshot(0)
	if err != nil {
		return
	}

	tree, err := ss.GetTree(t)
	if err != nil {
		return
	}

	err = tree.Delete(key)
	if err != nil {
		return
	}

	_, err = graviton.Commit(tree)
	if err != nil {
		return
	}

	return
}

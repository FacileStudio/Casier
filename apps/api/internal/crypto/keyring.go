package crypto

import "sync"

// Keyring holds the root encryption key shared by every service in the process.
// Root key rotation has to be visible to all of them at once: a service holding
// its own copy would keep decrypting with a key the database no longer matches.
type Keyring struct {
	mutex sync.RWMutex
	key   []byte
}

func NewKeyring(key []byte) *Keyring {
	return &Keyring{key: key}
}

func (keyring *Keyring) Key() []byte {
	keyring.mutex.RLock()
	defer keyring.mutex.RUnlock()
	return keyring.key
}

func (keyring *Keyring) Replace(key []byte) {
	keyring.mutex.Lock()
	defer keyring.mutex.Unlock()
	keyring.key = key
}

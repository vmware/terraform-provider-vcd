package vcd

import (
	"log"
	"sync"
)

// Imported from Hashicorp (https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html)

// mutexKV is a simple key/value store for arbitrary mutexes. It can be used to
// serialize changes across arbitrary collaborators that share knowledge of the
// keys they must serialize on.
//
// The initial use case is to let aws_security_group_rule resources serialize
// their access to individual security groups based on SG ID.
type mutexKV struct {
	lock   sync.Mutex
	store  map[string]*sync.Mutex
	silent bool
}

// Locks the mutex for the given key. Caller is responsible for calling kvUnlock
// for the same key
func (m *mutexKV) kvLock(key string) {
	if !m.silent {
		log.Printf("[DEBUG] Locking %q", key)
	}
	m.get(key).Lock()
	if !m.silent {
		log.Printf("[DEBUG] Locked %q", key)
	}
}

// kvUnlock the mutex for the given key. Caller must have called kvLock for the same key first
func (m *mutexKV) kvUnlock(key string) {
	if !m.silent {
		log.Printf("[DEBUG] Unlocking %q", key)
	}
	m.get(key).Unlock()
	if !m.silent {
		log.Printf("[DEBUG] Unlocked %q", key)
	}
}

// Returns a mutex for the given key, no guarantee of its lock status
func (m *mutexKV) get(key string) *sync.Mutex {
	m.lock.Lock()
	defer m.lock.Unlock()
	mutex, ok := m.store[key]
	if !ok {
		mutex = &sync.Mutex{}
		m.store[key] = mutex
	}
	return mutex
}

// Returns a properly initalized mutexKV
func newMutexKV() *mutexKV {
	return &mutexKV{
		store: make(map[string]*sync.Mutex),
	}
}

//  newMutexKVSilent returns a properly initalized mutexKV with the silent property set
func newMutexKVSilent() *mutexKV {
	return &mutexKV{
		store:  make(map[string]*sync.Mutex),
		silent: true,
	}
}

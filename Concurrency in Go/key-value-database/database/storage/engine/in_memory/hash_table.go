package in_memory

import "sync"

type HashTable struct {
	hmap map[string]string
	mx   sync.RWMutex
}

func NewHashTable() *HashTable {
	return &HashTable{
		hmap: make(map[string]string),
	}
}

func (ht *HashTable) Set(key string, value string) {
	ht.mx.Lock()
	defer ht.mx.Unlock()

	ht.hmap[key] = value
}

func (ht *HashTable) Get(key string) (value string, found bool) {
	ht.mx.RLock()
	defer ht.mx.RUnlock()

	value, found = ht.hmap[key]
	return
}

func (ht *HashTable) Del(key string) {
	ht.mx.Lock()
	defer ht.mx.Unlock()

	delete(ht.hmap, key)
}

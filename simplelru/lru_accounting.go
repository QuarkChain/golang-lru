package simplelru

import (
	"container/list"
	"errors"
)

// EvictCallback is used to get a callback when a cache entry is evicted

type AccountCallback func(key interface{}, value interface{}) int

// LRU implements a non-thread safe fixed size LRU cache
type LRUWithAccounting struct {
	limit     int
	size      int
	evictList *list.List
	items     map[interface{}]*list.Element
	onEvict   EvictCallback
	onAccount AccountCallback
}

// NewLRU constructs an LRU of the given size
func NewLRUWithAccounting(limit int, onAccount AccountCallback, onEvict EvictCallback) (*LRUWithAccounting, error) {
	if limit <= 0 {
		return nil, errors.New("must provide a positive size")
	}
	c := &LRUWithAccounting{
		limit:     limit,
		evictList: list.New(),
		items:     make(map[interface{}]*list.Element),
		onEvict:   onEvict,
		onAccount: onAccount,
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *LRUWithAccounting) Purge() {
	for k, v := range c.items {
		if c.onEvict != nil {
			c.onEvict(k, v.Value.(*entry).value)
		}
		delete(c.items, k)
	}
	c.evictList.Init()
	c.size = 0
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *LRUWithAccounting) Add(key, value interface{}) (evicted bool) {
	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		c.size -= c.onAccount(ent.Value.(*entry).key, ent.Value.(*entry).value)
		ent.Value.(*entry).value = value
		c.size += c.onAccount(ent.Value.(*entry).key, ent.Value.(*entry).value)

		return c.evictIfNeeded()
	}

	// Add new item
	ent := &entry{key, value}
	entry := c.evictList.PushFront(ent)
	c.items[key] = entry
	c.size += c.onAccount(key, value)

	return c.evictIfNeeded()
}

func (c *LRUWithAccounting) evictIfNeeded() (evicted bool) {
	evict := c.size > c.limit

	for c.size > c.limit {
		c.removeOldest()
	}

	return evict
}

// Get looks up a key's value from the cache.
func (c *LRUWithAccounting) Get(key interface{}) (value interface{}, ok bool) {
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		if ent.Value.(*entry) == nil {
			return nil, false
		}
		return ent.Value.(*entry).value, true
	}
	return
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRUWithAccounting) Contains(key interface{}) (ok bool) {
	_, ok = c.items[key]
	return ok
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *LRUWithAccounting) Peek(key interface{}) (value interface{}, ok bool) {
	var ent *list.Element
	if ent, ok = c.items[key]; ok {
		return ent.Value.(*entry).value, true
	}
	return nil, ok
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *LRUWithAccounting) Remove(key interface{}) (present bool) {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

// RemoveOldest removes the oldest item from the cache.
func (c *LRUWithAccounting) RemoveOldest() (key, value interface{}, ok bool) {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
		kv := ent.Value.(*entry)
		return kv.key, kv.value, true
	}
	return nil, nil, false
}

// GetOldest returns the oldest entry
func (c *LRUWithAccounting) GetOldest() (key, value interface{}, ok bool) {
	ent := c.evictList.Back()
	if ent != nil {
		kv := ent.Value.(*entry)
		return kv.key, kv.value, true
	}
	return nil, nil, false
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRUWithAccounting) Keys() []interface{} {
	keys := make([]interface{}, len(c.items))
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*entry).key
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *LRUWithAccounting) Len() int {
	return c.evictList.Len()
}

// Resize changes the cache size.
func (c *LRUWithAccounting) Resize(size int) (evicted int) {
	diff := c.Len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		c.removeOldest()
	}
	c.limit = size
	return diff
}

// removeOldest removes the oldest item from the cache.
func (c *LRUWithAccounting) removeOldest() {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

// AccountingSize returns the size of the cache measured by accounting func.
func (c *LRUWithAccounting) AccountingSize() int {
	return c.size
}

// removeElement is used to remove a given list element from the cache
func (c *LRUWithAccounting) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	kv := e.Value.(*entry)
	delete(c.items, kv.key)
	if c.onEvict != nil {
		c.onEvict(kv.key, kv.value)
	}
	c.size -= c.onAccount(kv.key, kv.value)
}

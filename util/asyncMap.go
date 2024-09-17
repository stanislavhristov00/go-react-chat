package util

import "sync"

type ThreadSafeMap[K comparable, V any] struct {
	objMap map[K]V
	lock   sync.RWMutex
}

func NewThreadSafeMap[K comparable, V any]() *ThreadSafeMap[K, V] {
	return &ThreadSafeMap[K, V]{
		objMap: make(map[K]V),
	}
}

func (self *ThreadSafeMap[K, V]) Get(key K) (V, bool) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	val, ok := self.objMap[key]
	return val, ok
}

func (self *ThreadSafeMap[K, V]) Set(key K, val V) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.objMap[key] = val
}

func (self *ThreadSafeMap[K, V]) Delete(key K) {
	self.lock.Lock()
	defer self.lock.Unlock()

	delete(self.objMap, key)
}

func (self *ThreadSafeMap[K, V]) SetAll(val V) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for key := range self.objMap {
		self.objMap[key] = val
	}
}

func (self *ThreadSafeMap[K, V]) RemoveAllElements() {
	self.lock.Lock()
	defer self.lock.Unlock()

	for key := range self.objMap {
		delete(self.objMap, key)
	}
}

package cache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	mu       sync.Mutex
	capacity int
	queue    List
	items    map[Key]*ListItem
}

type cacheItem struct {
	key   string
	value interface{}
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (l *lruCache) Set(key Key, value interface{}) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.items[key]
	if ok {
		l.queue.MoveToFront(item)
		item.Value = cacheItem{key: string(key), value: value}
		return true
	}

	i := cacheItem{key: string(key), value: value}
	l.queue.PushFront(i)
	l.items[key] = l.queue.Front()

	if len(l.items) > l.capacity {
		out := l.queue.Back()
		l.queue.Remove(out)
		delete(l.items, Key(out.Value.(cacheItem).key))
	}

	return false
}

func (l *lruCache) Get(key Key) (interface{}, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.items[key]
	if !ok {
		return nil, false
	}

	l.queue.MoveToFront(item)

	return item.Value.(cacheItem).value, true
}

func (l *lruCache) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	item := l.queue.Back()
	for {
		delete(l.items, Key(item.Value.(cacheItem).key))
		item = item.Next
		if item == nil {
			break
		}
	}
	l.queue = NewList()
}

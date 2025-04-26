package main

import (
	"container/list"
	"fmt"
	"sync"
)

// создание LruCache, представляющего кеш LRU
type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mutex    sync.Mutex
}

// создание записи кеша, представляющей запись в кеше LRU
type CacheEntry struct {
	key   string
	value interface{}
}

// NewLRUCache создает новый экземпляр LruCache с указанной ёмкостью
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// извлечение значения, связанного с данным ключом, из кеша
func (lru *LRUCache) Get(key string) (interface{}, bool) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if element, ok := lru.cache[key]; ok {
		// перемещение доступного элемента в начало списка (последний раз использованный)
		lru.list.MoveToFront(element)
		return element.Value.(*CacheEntry).value, true
	}

	return nil, false
}

// добавление или обновление пары ключ-значение в кеше
func (lru *LRUCache) Set(key string, value interface{}) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	// проверка того, существует ли ключ уже в кэше
	if element, ok := lru.cache[key]; ok {
		// обновляем существующую запись и перемещаем ее в начало (использовалась в последний раз)
		element.Value.(*CacheEntry).value = value
		lru.list.MoveToFront(element)
	} else {
		// добавление новой записи в кеш
		entry := &CacheEntry{key: key, value: value}
		element := lru.list.PushFront(entry)
		lru.cache[key] = element

		// проверяем, заполнен ли кеш (есть ли место), при необходимости удаляем наименее недавно использованный элемент
		if lru.list.Len() > lru.capacity {
			oldest := lru.list.Back()
			if oldest != nil {
				delete(lru.cache, oldest.Value.(*CacheEntry).key)
				lru.list.Remove(oldest)
			}
		}
	}
}

// PrintCache, который печатает текущее содержимое кеша
func (lru *LRUCache) PrintCache() {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	fmt.Printf("LRU Cache (Capacity: %d, Size: %d): [", lru.capacity, lru.list.Len())
	for element := lru.list.Front(); element != nil; element = element.Next() {
		entry := element.Value.(*CacheEntry)
		fmt.Printf("(%s: %v) ", entry.key, entry.value)
	}
	fmt.Println("]")
}

func main() {
	// создание кэша LRU ёмкостью 3
	lruCache := NewLRUCache(3)

	// задаём пары ключ-занчение
	lruCache.Set("company", "Yandex")
	lruCache.Set("division", "Yandex Lyceum")
	lruCache.Set("course", "Golang")

	lruCache.PrintCache()

	if value, ok := lruCache.Get("company"); ok {
		fmt.Println("Value for company:", value)
	} else {
		fmt.Println("company not found in the cache.")
	}

	// установка дополнительных пар ключ-значение для инициирования вытеснения
	lruCache.Set("year", 2024)
	lruCache.Set("age", "13-17yrs")

	lruCache.PrintCache()
}

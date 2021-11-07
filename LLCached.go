package LLCached

import (
	"fmt"
	"log"
	"sync"
)

type LL interface {
	Get(key string) ([]byte, error)
}

type Group struct {
	name      string
	ll        LL
	mainCache cache
}

func (f LLFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type LLFunc func(key string) ([]byte, error)

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[LL] hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.ll.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func NewGroup(name string, cacheBytes int64, ll LL) *Group {
	if ll == nil {
		panic("nil LL")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		ll:        ll,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

package llcache

import (
	"fmt"
	"log"
	"sync"

	pb "github.com/go-ll/llcache/llcachepb"
	"github.com/go-ll/llcache/singleflight"
)

type LL interface {
	Get(key string) ([]byte, error)
}

type Group struct {
	name      string
	ll        LL
	mainCache cache
	peers     PeerPicker
	loader    *singleflight.Group
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
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[llcache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
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
		loader:    &singleflight.Group{},
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

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) getFromPeer(peer PeerLL, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

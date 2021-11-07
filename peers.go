package LLCached

import (
	"github.com/cexll/LLCached/consistenthash"
	pb "github.com/cexll/LLCached/llcachepb"
)

type PeerPicker interface {
	PickPeer(key string) (peer PeerLL, ok bool)
}

type PeerLL interface {
	Get(in *pb.Request, out *pb.Response) error
}

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpLLs = make(map[string]*httpLL, len(peers))
	for _, peer := range peers {
		p.httpLLs[peer] = &httpLL{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerLL, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpLLs[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)

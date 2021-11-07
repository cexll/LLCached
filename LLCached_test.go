package LLCached

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"testing"
)

var db = map[string]string{
	"Tom":  "530",
	"Jack": "123",
	"Sam":  "231",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	ll := NewGroup("scores", 2<<10, LLFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	for k, v := range db {
		if view, err := ll.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		}
		if _, err := ll.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := ll.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty but %s got", view)
	}
}

func createGroup() *Group {
	return NewGroup("scores", 2<<10, LLFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db1[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, g *Group) {
	peers := NewHTTPPool(addr)
	peers.Set(addrs...)
	g.RegisterPeers(peers)
	log.Println("LLCached is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, g *Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := g.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func TestMain(t *testing.T) {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "LLCached server port")
	flag.BoolVar(&api, "api", false, "Start a api server")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	g := createGroup()
	if api {
		go startAPIServer(apiAddr, g)
	}
	startCacheServer(addrMap[port], []string(addrs), g)
}

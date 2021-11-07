package LLCached

import (
	"fmt"
	"log"
	"net/http"
	"testing"
)

var db1 = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestHTTP(t *testing.T) {
	NewGroup("scores", 2<<10, LLFunc(
		func(key string) ([]byte, error) {
			log.Println("[Slowdb1] search key", key)
			if v, ok := db1[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := NewHTTPPool(addr)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}

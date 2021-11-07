package LLCached

import (
	"fmt"
	"log"
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

package main

import (
	"fmt"
	"github.com/Kingpie/kinCache/cache"
	"log"
	"net/http"
)

var db = map[string]string{
	"hehe": "1",
	"haha": "2",
	"lala": "3",
}

func main() {
	cache.NewGroup("rank", 1024, cache.GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "127.0.0.1:9527"
	pool := cache.NewHTTPPool(addr)
	log.Fatal(http.ListenAndServe(addr, pool))
}

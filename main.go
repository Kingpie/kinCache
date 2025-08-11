package main

import (
	"flag"
	"fmt"
	"kinCache/cache"
	"log"
	"net/http"
)

var db = map[string]string{
	"hehe": "1",
	"haha": "2",
	"lala": "3",
	"lili": "4",
	"cici": "5",
	"gaga": "6",
}

func createGroup() *cache.Group {
	return cache.NewGroup("rank", 1024, cache.GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startServer(addr string, addrs []string, group *cache.Group) {
	server := cache.NewHTTPPool(addr)
	server.Set(addrs...)
	group.RegisterPeers(server)
	log.Println("cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], server))
}

// api server
func startAPIServer(apiAddr string, group *cache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			data, err := group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(data.Copy())
		}))
	log.Println("api server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "cache server port")
	flag.BoolVar(&api, "api", false, "start an api server?")
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

	group := createGroup()
	if api {
		go startAPIServer(apiAddr, group)
	}

	startServer(addrMap[port], addrs, group)
}

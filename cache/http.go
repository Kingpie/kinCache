package cache

import (
	"fmt"
	"github.com/Kingpie/kinCache/consistent_hash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	defaultPath     = "/_kincache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	addr        string //本机地址
	basePath    string //默认前缀
	mtx         sync.Mutex
	peersMap    *consistent_hash.Map //节点map
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(addr string) *HTTPPool {
	return &HTTPPool{
		addr:     addr,
		basePath: defaultPath,
	}
}

func (p *HTTPPool) Set(peers ...string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.peersMap = consistent_hash.New(defaultReplicas, nil)
	p.peersMap.AddNode(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if peer := p.peersMap.Get(key); peer != "" && peer != p.addr {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s %s] %s", p.addr, time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		p.Log("Invalid Url:%s", r.URL.Path)
		return
	}
	p.Log("req %s %s", r.Method, r.URL.Path)

	//<basepath>/<group>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}

	data, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-steam")
	w.Write(data.Copy())
}

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))

	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body:%v", err)
	}

	return bytes, nil
}

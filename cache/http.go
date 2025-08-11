package cache

import (
	"fmt"
	"io"
	"kinCache/consistent_hash"
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

// Set updates the pool's peer list and rebuilds the consistent hash ring.
// It takes a variadic list of peer addresses to set as the new peer list.
//
// Parameters:
//   - peers: a variadic list of string addresses representing the peer nodes
//
// This function is thread-safe and will replace the existing peer list
// with the new one, rebuilding the consistent hash ring and HTTP getters.
func (p *HTTPPool) Set(peers ...string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	// Rebuild the consistent hash ring with the new peer list
	p.peersMap = consistent_hash.New(defaultReplicas, nil)
	p.peersMap.AddNode(peers...)

	// Create new HTTP getters for each peer
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer 根据给定的key选择一个合适的peer节点
// 参数:
//
//	key: 用于选择peer的键值
//
// 返回值:
//
//	PeerGetter: 选中的peer节点的getter接口，如果未找到合适的peer则返回nil
//	bool: 如果成功找到并返回了peer节点则为true，否则为false
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	// 根据key获取对应的peer地址，如果peer存在且不是当前节点自身，则返回该peer的getter
	if peer := p.peersMap.GetNode(key); peer != "" && peer != p.addr {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s %s] %s", p.addr, time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, v...))
}

// ServeHTTP 实现了http.Handler接口，处理HTTP请求
// w: HTTP响应写入器，用于向客户端返回数据
// r: HTTP请求对象，包含客户端请求信息
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 检查请求路径是否以basePath开头
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		p.Log("Invalid Url:%s", r.URL.Path)
		return
	}
	p.Log("req %s %s", r.Method, r.URL.Path)

	// 解析URL路径，格式为<basepath>/<group>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	// 根据组名获取对应的group对象
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}

	// 从group中获取指定key的数据
	data, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置响应头并返回数据
	w.Header().Set("Content-Type", "application/octet-steam")
	w.Write(data.Copy())
}

type httpGetter struct {
	baseURL string
}

// Get 从远程服务器获取指定组和键的数据
// 参数:
//
//	group - 数据组名
//	key - 数据键名
//
// 返回值:
//
//	[]byte - 获取到的数据字节切片
//	error - 获取过程中发生的错误
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	// 构造完整的请求URL，包含baseURL和经过URL编码的group、key
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))

	// 发送HTTP GET请求
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	// 确保响应体在函数结束时关闭
	defer res.Body.Close()

	// 检查HTTP响应状态码，非200状态码视为错误
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %v", res.Status)
	}

	// 读取响应体中的数据
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body:%v", err)
	}

	return bytes, nil
}

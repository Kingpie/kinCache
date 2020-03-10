package cache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const defaultPath = "/_kincache/"

type HTTPPool struct {
	addr     string
	basePath string
}

func NewHTTPPool(addr string) *HTTPPool {
	return &HTTPPool{
		addr:     addr,
		basePath: defaultPath,
	}
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

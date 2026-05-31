package cache

import (
	"log"
	"net/http"
	"sync"
)

/*
todo:
 improve key-generating (for auth-security)
 add timeout and ttl for keys
 add lru
*/

type Cache struct {
	mu sync.RWMutex
	m  map[string]CachedResponse
}

type CachedResponse struct {
	Code   int
	Header http.Header
	Body   []byte
}

func NewCache() *Cache {
	return &Cache{m: make(map[string]CachedResponse)}
}

func (c *Cache) Get(r *http.Request) (res CachedResponse, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res, ok = c.m[key(r)]
	log.Println("Gone to cache:", res, ok)
	return res, ok
}

func (c *Cache) Set(req *http.Request, res *http.Response, body []byte) (cr CachedResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cr = CachedResponse{
		Code:   res.StatusCode,
		Header: res.Header.Clone(),
		Body:   body,
	}
	c.m[key(req)] = cr
	return cr
}

func (c *CachedResponse) ToHttpResponse() *http.Response {
	return &http.Response{
		StatusCode: c.Code,
		Header:     c.Header,
	}
}

func key(r *http.Request) string {
	return r.Method + "|" + r.URL.Path + "?" + r.URL.RawQuery
}

package proxy

import (
	"caching-proxy/internal/cache"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Proxy struct {
	target *url.URL
	cache  *cache.Cache
	client *http.Client
}

func NewServer(port string, target *url.URL) http.Server {
	p := newProxy(target)

	return http.Server{
		Addr:    "localhost:" + port,
		Handler: http.HandlerFunc(p.handle),
	}
}

func parseUrl(base *url.URL, r *url.URL) *url.URL {
	reqURL := &url.URL{
		Path:     r.Path,
		RawQuery: r.RawQuery,
	}
	target := base.ResolveReference(reqURL)
	return target
}

func sendError(w http.ResponseWriter, e error) {
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte(e.Error()))
	if err != nil {
		log.Println("Error writing response:", err)
		return
	}
	return
}

func (p *Proxy) handle(w http.ResponseWriter, r *http.Request) {
	// todo: add timeout + `done`
	targetURL := parseUrl(p.target, r.URL)

	req, err := http.NewRequestWithContext(
		r.Context(),
		r.Method,
		targetURL.String(),
		r.Body,
	)
	if err != nil {
		log.Println("Error creating request:", err)
		sendError(w, err)
	}
	req.Header = make(http.Header)
	copyHeader(req.Header, r.Header)

	val, ok := p.cache.Get(req)
	if !ok {
		val, err = p.handleCacheMiss(w, req)
	}
	p.sendCacheHit(w, val)

}

func (p *Proxy) sendCacheHit(w http.ResponseWriter, r cache.CachedResponse) {
	copyHeader(w.Header(), r.Header)
	w.WriteHeader(r.Code)
	_, err := w.Write(r.Body)
	if err != nil {
		log.Println("cannot write response body: " + err.Error())
	}
}

func (p *Proxy) handleCacheMiss(w http.ResponseWriter, req *http.Request) (cache.CachedResponse, error) {
	proxyResponse, err := p.client.Do(req)
	if err != nil {
		log.Println("cannot make request: " + err.Error())
		sendError(w, err)
		return cache.CachedResponse{}, err
	}
	defer func() {
		if err := proxyResponse.Body.Close(); err != nil {
			log.Println("cannot close response body: " + err.Error())
		}
	}()

	body, err := io.ReadAll(proxyResponse.Body)
	if err != nil {
		log.Println("cannot read response body: " + err.Error())
		sendError(w, err)
		return cache.CachedResponse{}, err
	}

	return p.cache.Set(req, proxyResponse, body), nil
}

func newProxy(target *url.URL) *Proxy {
	return &Proxy{
		target: target,
		cache:  cache.NewCache(),
		client: http.DefaultClient,
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

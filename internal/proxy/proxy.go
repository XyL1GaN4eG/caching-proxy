package proxy

import (
	"caching-proxy/internal/cache"
	"caching-proxy/internal/util"
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
	targetURL := util.ParseURL(p.target, r.URL)

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

	val, ok := p.cache.Get(req)
	if !ok {
		val, err = p.handleCacheMiss(w, req)
		if err != nil {
			log.Println("Error caching response:", err)
			return
		}
		headers := val.Header.Clone()
		headers.Set("X-Cache", "MISS")
		copyHeader(w.Header(), headers)
	} else {
		headers := val.Header.Clone()
		headers.Set("X-Cache", "HIT")
		copyHeader(w.Header(), headers)
	}
	p.send(w, val)
}

func (p *Proxy) send(w http.ResponseWriter, r cache.CachedResponse) {
	w.WriteHeader(r.Code)
	_, err := w.Write(r.Body)
	if err != nil {
		log.Println("cannot write response body: " + err.Error())
	}
}

// fixme: убрать w http.ResponseWriter, чтобы чисто ошибка возвращалась
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
	resp := p.cache.Set(req, proxyResponse, body)
	return resp, nil
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

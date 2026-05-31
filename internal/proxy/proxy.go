package proxy

import (
	"caching-proxy/internal/cache"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Proxy struct {
	target string
	cache  *cache.Cache
	client *http.Client
}

func NewServer(port, target string) http.Server {
	p := newProxy(target)

	return http.Server{
		Addr:    "localhost:" + port,
		Handler: http.HandlerFunc(p.handle),
	}
}

func (p *Proxy) handle(w http.ResponseWriter, r *http.Request) {
	// todo: add timeout + `done`
	log.Printf("method:%v", r.Method)
	log.Printf("target:%v\n", p.target)
	log.Printf("path:%v\n", r.URL.Path)
	log.Printf("query:%v\n", r.URL.RawQuery)
	base, err := url.Parse(p.target)
	if err != nil {
		panic(err)
	}
	reqURL := &url.URL{
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	targetURL := base.ResolveReference(reqURL)
	log.Printf("full path:%v\n", targetURL)

	req, _ := http.NewRequestWithContext(
		r.Context(),
		r.Method,
		targetURL.String(),
		r.Body,
	)
	req.Header = make(http.Header)
	copyHeader(req.Header, r.Header)

	val, ok := p.cache.Get(req)
	var proxyResponse *http.Response
	var body []byte
	if !ok {
		proxyResponse, err = p.client.Do(req)
		if err != nil {
			log.Println("cannot make request: " + err.Error())
			return
		}
		defer proxyResponse.Body.Close()

		body, err = io.ReadAll(proxyResponse.Body)
		if err != nil {
			log.Println("cannot read response body: " + err.Error())
		}
		log.Println("BODY: " + string(body))
		p.cache.Set(req, proxyResponse, body)
	} else {
		proxyResponse = val.ToHttpResponse()
		body, err = io.ReadAll(proxyResponse.Body)
		if err != nil {
			log.Println("cannot read response body: " + err.Error())
		}
		defer proxyResponse.Body.Close()
	}
	copyHeader(w.Header(), proxyResponse.Header)

	w.WriteHeader(proxyResponse.StatusCode)
	_, err = w.Write(body)
	if err != nil {
		log.Println("cannot write body: " + err.Error())
		return
	}
}

func newProxy(target string) *Proxy {
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

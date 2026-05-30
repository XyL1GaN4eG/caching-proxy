package proxy

import (
	"caching-proxy/internal/cache"
	"io"
	"io/ioutil"
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
	if !ok {
		proxyResponse, err = p.client.Do(req)
		if err != nil {
			log.Println("cannot make request: " + err.Error())
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Println("cannot close body: " + err.Error())
			}
		}(proxyResponse.Body)
		body, err := ioutil.ReadAll(proxyResponse.Body) // fixme
		if err != nil {
			log.Println("cannot read response body: " + err.Error())
		}
		p.cache.Set(req, proxyResponse, body)
	} else {
		proxyResponse = val.ToHttpResponse()
	}
	copyHeader(w.Header(), proxyResponse.Header)

	w.WriteHeader(proxyResponse.StatusCode)
	_, err = io.Copy(w, proxyResponse.Body)
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

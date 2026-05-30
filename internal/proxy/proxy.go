package proxy

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

type Proxy struct {
	target string
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

	get, err := p.client.Do(req)
	if err != nil {
		log.Println("cannot make request: " + err.Error())
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("cannot close body: " + err.Error())
		}
	}(get.Body)

	copyHeader(w.Header(), get.Header)

	w.WriteHeader(get.StatusCode)
	_, err = io.Copy(w, get.Body)
	if err != nil {
		log.Println("cannot write body: " + err.Error())
		return
	}
}

func newProxy(target string) *Proxy {
	return &Proxy{
		target: target,
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

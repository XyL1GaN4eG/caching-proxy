package main

import (
	"caching-proxy/internal/flags"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

func main() {
	port, target := flags.Parse()
	fmt.Println("port:", port)
	fmt.Println("target:", target)
	p := newProxy(target)
	s := http.Server{
		Addr:    "localhost:" + port,
		Handler: http.HandlerFunc(p.handle),
	}
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

type Proxy struct {
	target string
	client *http.Client
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
		panic("cannot get: " + err.Error())
	}

	copyHeader(w.Header(), get.Header)

	w.WriteHeader(get.StatusCode)
	_, err = io.Copy(w, get.Body)
	if err != nil {
		panic("cannot write body: " + err.Error())
	}
	err = get.Body.Close()
	if err != nil {
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

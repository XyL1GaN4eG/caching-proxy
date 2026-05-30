package main

import (
	"caching-proxy/internal/flags"
	"fmt"
	"io"
	"log"
	"net/http"
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
	// todo: add timeout
	log.Printf("method:%v", r.Method)
	log.Printf("target:%v\n", p.target)
	log.Printf("path:%v\n", r.URL.Path)
	log.Printf("query:%v\n", r.URL.RawQuery)
	var targetUrl string
	if r.URL.RawQuery == "" {
		targetUrl = p.target + r.URL.Path
	} else {
		targetUrl = p.target + r.URL.Path + "?" + r.URL.RawQuery
	}
	log.Printf("full path:%v\n", targetUrl)

	req, _ := http.NewRequest(r.Method, targetUrl, r.Body)
	for k, v := range r.Header {
		req.Header.Set(k, v[0])
	}

	get, err := p.client.Do(req)
	if err != nil {
		panic("cannot get: " + err.Error())
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(get.Body)

	body, err := io.ReadAll(get.Body)
	if err != nil {
		panic("cannot read body: " + err.Error())
	}
	_, err = w.Write(body)
	if err != nil {
		panic("cannot write body: " + err.Error())
	}
}

func newProxy(target string) *Proxy {
	return &Proxy{
		target: target,
		client: http.DefaultClient,
	}
}

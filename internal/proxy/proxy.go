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

func parseUrl(base *url.URL, r *url.URL) (*url.URL, error) {
	reqURL := &url.URL{
		Path:     r.Path,
		RawQuery: r.RawQuery,
	}
	target := base.ResolveReference(reqURL)
	return target, nil
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
	// log.Printf("method:%v", r.Method)
	// log.Printf("target:%v\n", p.target)
	// log.Printf("path:%v\n", r.URL.Path)
	// log.Printf("query:%v\n", r.URL.RawQuery)

	targetURL, err := parseUrl(p.target, r.URL)
	if err != nil {
		log.Println("Error parsing URL:", err)
		sendError(w, err)
		return
	}

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
			sendError(w, err)
			return
		}
		defer func() {
			err = proxyResponse.Body.Close()
			if err != nil {
				log.Println("cannot close response body: " + err.Error())
				sendError(w, err)
				return
			}
		}()

		body, err = io.ReadAll(proxyResponse.Body)
		if err != nil {
			log.Println("cannot read response body: " + err.Error())
			sendError(w, err)
			return
		}

		p.cache.Set(req, proxyResponse, body)
	} else {
		proxyResponse = val.ToHttpResponse()
		body = val.Body
	}
	copyHeader(w.Header(), proxyResponse.Header)

	w.WriteHeader(proxyResponse.StatusCode)
	_, err = w.Write(body)
	if err != nil {
		log.Println("cannot write body: " + err.Error())
		return
	}
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

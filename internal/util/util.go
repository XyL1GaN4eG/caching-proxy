package util

import "net/url"

func ParseURL(base *url.URL, r *url.URL) *url.URL {
	reqURL := &url.URL{
		Path:     r.Path,
		RawQuery: r.RawQuery,
	}
	target := base.ResolveReference(reqURL)
	return target
}

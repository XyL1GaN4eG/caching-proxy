package flags

import (
	"flag"
	"log"
	"net/url"
	"strconv"
)

func Parse() (port string, target *url.URL) {
	portPtr := flag.Uint("port", 8080, "a port number")
	targetPtr := flag.String("target", "localhost", "a target url")
	flag.Parse()
	if !parseUrl(targetPtr) {
		log.Fatal("Invalid target url")
	}
	v := *targetPtr
	target, err := url.Parse(v)
	if err != nil {
		log.Fatal("Invalid target url")
	}
	return strconv.Itoa(int(*portPtr)), target
}

func parseUrl(_ *string) bool {
	// todo: use regexp + normalize url
	return true
}

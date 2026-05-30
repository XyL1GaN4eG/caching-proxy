package flags

import (
	"flag"
	"log"
	"strconv"
)

func Parse() (port, target string) {
	portPtr := flag.Uint("port", 8080, "a port number")
	targetPtr := flag.String("target", "localhost", "a target url")
	flag.Parse()
	if !parseUrl(targetPtr) {
		log.Fatal("Invalid target url")
	}
	return strconv.Itoa(int(*portPtr)), *targetPtr
}

func parseUrl(_ *string) bool {
	// todo: use regexp + normalize url
	return true
}

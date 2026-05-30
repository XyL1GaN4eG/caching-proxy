package main

import "flag"
import "fmt"

func main() {
	port, origin := parseArgs()
	fmt.Println("port:", port)
	fmt.Println("origin:", origin)
}

func parseArgs() (port uint16, origin string) {
	portPtr := flag.Uint("port", 8080, "a port number")
	originPtr := flag.String("origin", "http://localhost", "a origin url")
	flag.Parse()
	return uint16(*portPtr), *originPtr
}

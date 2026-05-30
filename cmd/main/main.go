package main

import (
	"caching-proxy/internal/flags"
	"caching-proxy/internal/proxy"
	"fmt"
)

func main() {
	port, target := flags.Parse()
	fmt.Println("port:", port)
	fmt.Println("target:", target)
	s := proxy.NewServer(port, target)
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

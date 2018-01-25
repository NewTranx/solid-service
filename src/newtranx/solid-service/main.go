package main

import (
	"flag"
	"newtranx/solid-service/server"
)

func main() {
	server := &server.ServiceEndpoint{}
	flag.StringVar(&server.Host, "h", "", "host")
	flag.IntVar(&server.Port, "p", 8080, "port")
	flag.StringVar(&server.WorkPath, "w", "./work", "work path")
	flag.Parse()
	server.Start()
}

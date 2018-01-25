package main

import (
	"flag"
	"newtranx/solid-service/server"
	"os"
	"fmt"
)

func main() {
	server := &server.ServiceEndpoint{}
	flag.StringVar(&server.Host, "h", "", "host")
	flag.IntVar(&server.Port, "p", 8080, "port")
	flag.StringVar(&server.WorkPath, "w", ".", "work path")
	flag.BoolVar(&server.Cleanup, "no-cleanup", false, "disable cleanup after conversion")
	flag.Parse()
	server.Cleanup = !server.Cleanup
	initWorkPath(server.WorkPath)
	server.Start()
}

func initWorkPath(path string) {
	ensurePath(path + "/" + server.SrcPath)
	ensurePath(path + "/" + server.OutputPath)
	ensurePath(path + "/" + server.ErrPath)
}

func ensurePath(path string) {
	if info, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			panic(err)
		}
	} else if err == nil {
		if !info.IsDir() {
			panic(fmt.Errorf("%s is not a folder", path))
		}
	} else {
		panic(err)
	}
}

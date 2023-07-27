package main

import (
	"encoding/gob"
	"github.com/blang/mpv"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

func main() {
	sock := mpv.NewIPCClient("/tmp/mpvsocket")

	gob.Register(map[string]interface{}{})

	serve := mpv.NewRPCServer(sock)

	err := rpc.Register(serve)
	if err != nil {
		log.Fatalf("not in rpc format")
	}

	rpc.HandleHTTP()

	listener, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("Listen error: ", e)
	}
	log.Printf("Serving RPC server on port %d", 1234)

	http.Serve(listener, nil)
}

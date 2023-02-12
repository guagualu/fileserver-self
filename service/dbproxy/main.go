package main

import (
	pb "filestore/service/dbproxy/proto"
	dbproxy "filestore/service/dbproxy/rpc"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

func main() {
	Serve := grpc.NewServer()
	rg := &dbproxy.DBProxyStruct{}
	pb.RegisterDBProxyServiceServer(
		Serve, rg,
	)
	//todo :etcd 服务发现
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		fmt.Println("listen err:=", err)
		return
	}
	err = Serve.Serve(lis)
	if err != nil {
		fmt.Println("serve err:=", err)
		return
	}
}

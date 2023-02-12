package client

import (
	"context"
	"fmt"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

//一个标志
const Schema = "etcd"

//service 服务名 endpoints 服务注册中心节点
func NewClientConn(service string, endpoints []string, ctx context.Context) (*grpc.ClientConn, error) {
	//1、创建一个reslover 并将service注册进去
	r := NewReslover(endpoints, service)
	resolver.Register(r)
	timectx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//2、进行grpc的dail 可以使用超时等dailoption balancername好像被弃用了 超时使用dialwithcontext
	//感觉conn的超时只是在单次连接时 而不是对pb中的服务们整体来说
	conn, err := grpc.DialContext(timectx, service, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	//3、返回conn和err
	return conn, nil
}

//regesiter需要builder builder需要实现build 和scheme  同时为了使用etcd和resloverclientconn的函数
//同时还要实现resolver的resolvernow 和close
type Reslover struct {
	//服务注册中心j集群地址
	Endpoints []string
	//servicename
	ServiceName string
	//etcd客户端
	Cli *clientv3.Client
	//reslover的客户端连接
	Reslovercli resolver.ClientConn
}

//创建一个新的resolver
func NewReslover(endpoints []string, servicename string) resolver.Builder {

	return Reslover{Endpoints: endpoints, ServiceName: servicename}
}

//将在grpc.dail时被调用  target时dail时的url相关信息
func (r Reslover) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	//1、建立服务注册中心的连接

	//2、定义好要在etcd中查找的key
	prefix := "/" + target.URL.Scheme + "/" + r.ServiceName
	//3、启动watch携程 watch函数包括了state也就是address初始化以及更新操作   这同时说明不管哪的携程只要main不停 那么他就可以一直运行
	go r.watch(prefix)
	//4、返回
	return r, nil

}

//watch函数包括了state也就是address初始化以及更新操作
func (r Reslover) watch(perfix string) {
	//1、建立从服务注册中心获得的每个key对应的map（因为同一个服务的不同服务器存入etcd的key后缀不同）和updatestate的[]address
	addrDirect := make(map[string]resolver.Address)
	//2、etcd中查找的prefix 获得服务器地址列表
	res, err := r.Cli.Get(context.Background(), perfix, clientv3.WithPrefix())
	if err != nil {
		fmt.Println("etcd get err:", err)
		return
	}
	for _, v := range res.Kvs {
		addrDirect[string(v.Key)] = resolver.Address{Addr: string(v.Value)}
	}
	//3、定义update函数用于监听到服务器改动时的更新操作 与初始化一样
	update := func() {
		addrlist := make([]resolver.Address, len(addrDirect))
		err := r.Reslovercli.UpdateState(resolver.State{Addresses: addrlist})
		if err != nil {
			fmt.Println("update err:", err)
			return
		}
	}

	//4、初始化 将服务器list转为[]address中 然后调用update
	update()
	//5、调用etcd watch 对返回的watchchan进行range遍历 每当对应events时 进行更新操作
	watchchan := r.Cli.Watch(context.Background(), perfix, clientv3.WithPrefix())
	for wchan := range watchchan {
		for _, e := range wchan.Events {
			//如果检测到删除 delete dict对应
			if e.Type == mvccpb.DELETE {
				delete(addrDirect, string(e.Kv.Key))
			}
			if e.Type == mvccpb.PUT {
				addrDirect[string(e.Kv.Key)] = resolver.Address{Addr: string(e.Kv.Value)}
			}
		}
		update()
	}

}
func (r Reslover) Scheme() string {
	return Schema
}

//对传入进来的serviname进行解析 无必要不用
func (r Reslover) ResolveNow(rn resolver.ResolveNowOptions) {

}

//etcd客户端关闭
func (r Reslover) Close() {
	r.Cli.Close()
}

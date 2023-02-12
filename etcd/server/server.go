package server

import (
	"context"
	"fmt"

	// "github.com/coreos/etcd/clientv3"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const schema = "etcd"

type RegisterService struct {
	Client        *clientv3.Client
	LeaseID       clientv3.LeaseID                        //创建租约后leasegrantresponse 的id 他还包括了ttl和error
	KeepAliveChan <-chan *clientv3.LeaseKeepAliveResponse //responseheader id ；ttl等 当自动续约时会发送归来
	Key           string                                  //存入的key
	Value         string                                  //key 对应的value

}

type RegisterConfig struct {
	conf       clientv3.Config
	ServerName string //服务名
	Address    string //服务集群地址
	LeaseTTL   int64  //租约ttl

}

func NewServiceRegister(conf RegisterConfig) (*RegisterService, error) {
	//1、过conf new出新的client
	client, err := clientv3.New(conf.conf)
	if err != nil {
		fmt.Println("client new err:=", err)
		return nil, err
	}
	//2、对所需要注册的服务进行注册
	rgservice := &RegisterService{
		Client: client,
		Key:    "/" + schema + "/" + conf.ServerName + "/" + conf.Address,
		Value:  conf.Address,
	}
	//3、先grant 租约 ttl是conf传进来的

	//4、kv操作中的put 的】opoption采用绑定租约

	//5、开启keepalive  监听keepalivechan 保证续约成功
	//345交给绑定续约操作
	rgserver, err := rgservice.GrantLeaseAndPut(conf.LeaseTTL)
	if err != nil {
		fmt.Println("grantleaseandput err:=", err)
		return nil, err
	}
	return rgserver, nil
}

//创建租约 kv操作 设置keepalive
func (rg *RegisterService) GrantLeaseAndPut(LeaseTTl int64) (*RegisterService, error) {
	//3、先grant 租约 ttl是conf传进来的
	leaseResp, err := rg.Client.Grant(context.Background(), LeaseTTl)
	if err != nil {
		fmt.Println("grant err:", err)
		return nil, err
	}
	rg.LeaseID = leaseResp.ID
	//4、kv操作中的put 的】opoption采用绑定租约
	_, err = rg.Client.Put(context.Background(), rg.Key, rg.Value, clientv3.WithLease(rg.LeaseID))
	if err != nil {
		fmt.Println("put err:", err)
		return nil, err
	}
	//5、开启keepalive  监听keepalivechan 保证续约成功
	keepchan, err := rg.Client.KeepAlive(context.Background(), rg.LeaseID)
	if err != nil {
		fmt.Println("keepalive err:", err)
		return nil, err
	}
	rg.KeepAliveChan = keepchan
	return rg, nil
}

//撤销服务 撤销lease revoke  并且关闭cli连接
func (rg *RegisterService) Close() {
	_, err := rg.Client.Revoke(context.Background(), rg.LeaseID)
	if err != nil {
		fmt.Println("revoke err:", err)
		return
	}
	rg.Client.Close()
	return

}

//监听keepalive程序 将在携程中使用
func (rg *RegisterService) ListenKeepalivechan() {
	//1、接受接受传给chan的信息
	for k := range rg.KeepAliveChan {
		if k == nil {
			fmt.Println("续约失败")
		}
		fmt.Println("leaseid" + string(k.ID) + "完成续约")
	}

	//2、如果接受成功 打印 如果出错 打印续约失败
	fmt.Println("续约失败")
}

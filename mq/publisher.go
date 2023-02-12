package mq

import (
	"fmt"

	"github.com/streadway/amqp"
)

var (
	Rabconn    *amqp.Connection
	rabchannel *amqp.Channel
)

type MqFileInfo struct {
	FileHash    string
	FileName    string
	CurLocateAt string
}

func init() {
	var err error

	//1、获取mq连接
	Rabconn, err = amqp.Dial("amqp://guest:guest@127.0.0.1:5672/")
	if err != nil {
		fmt.Println("rabconn dail err:", err)
		return
	}
	//2、获取mq的channel
	rabchannel, err = Rabconn.Channel()
	if err != nil {
		fmt.Println("rabconn channel err:", err)
		return
	}

}

func Rabpublish(routekey, msg string) {
	//如果在web控制台创建了exchange 1可以不用  发布端只负责发布到交换机
	//1、先定义出exchange
	err := rabchannel.ExchangeDeclare( //??? 属性细看
		"filestore-ceph",
		"direct", //路由类型
		true,     //是否持久化
		false,    //是否丢失自动删除
		false,    //是否具有排他性
		false,    //是否堵塞
		nil,      //额外属性
	)
	if err != nil {
		fmt.Println("rabconn channel err:", err)
		return
	} else {
		//错误处理 只不过提前在web端控制台设置好就不会出现这个问题

	}

	//4、publish
	err = rabchannel.Publish(
		"filestore-ceph",
		routekey,
		true,  //如果为 ture 如果无法找到符合条件的队列 那么返回信息给发送者
		false, //不起作用？
		amqp.Publishing{
			ContentType: "text/plain", //明文格式
			Body:        []byte(msg),  //发送的信息 只能是字节信息
			Expiration:  "5000",       //ttl 这是5s
		})
	if err != nil {
		//错误处理 这个可以是转到死信交换机
		return
	}
	return

}

func Redispublish(routekey, msg string) {
	//如果在web控制台创建了exchange 1可以不用  发布端只负责发布到交换机
	//1、先定义出exchange
	err := rabchannel.ExchangeDeclare( //??? 属性细看
		"filestore-redis",
		"direct", //路由类型
		true,     //是否持久化
		false,    //是否丢失自动删除
		false,    //是否具有排他性
		false,    //是否堵塞
		nil,      //额外属性
	)
	if err != nil {
		fmt.Println("rabconn channel err:", err)
		return
	} else {
		//错误处理 只不过提前在web端控制台设置好就不会出现这个问题

	}

	//4、publish
	err = rabchannel.Publish(
		"filestore-redis",
		routekey,
		true,  //如果为 ture 如果无法找到符合条件的队列 那么返回信息给发送者
		false, //不起作用？
		amqp.Publishing{
			ContentType: "text/plain", //明文格式
			Body:        []byte(msg),  //发送的信息 只能是字节信息
			Expiration:  "5000",       //ttl 这是5s
		})
	if err != nil {
		//错误处理 这个可以是转到死信交换机
		return
	}
	return

}

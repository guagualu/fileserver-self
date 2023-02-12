package mq

import (
	"fmt"
	"log"
)

func RabConsumer(callback func(message []byte) error) error {
	//1、定义channel exchange 绑定 如果在控制台设置好 那么就不用
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
		return err
	} else {
		//错误处理 只不过提前在web端控制台设置好就不会出现这个问题

	}
	_, err = rabchannel.QueueDeclare(
		"filestore-channel-ceph",
		//是否持久化
		false,
		//是否自动删除
		false,
		//是否具有排他性 只能有一个消费者监听
		false,
		//是否堵塞
		false,
		nil, //额外属性  在这里是绑定死信机
	)
	if err != nil {
		fmt.Println("rabconn queue err:", err)
		return err
	} else {
		//错误处理 只不过提前在web端控制台设置好就不会出现这个问题

	}
	err = rabchannel.QueueBind(
		"filestore-channel-ceph",
		"ceph",
		"filestore-ceph",
		false,
		nil,
	)
	if err != nil {
		fmt.Println("rabconn bind err:", err)
		return err
	} else {
		//错误处理 只不过提前在web端控制台设置好就不会出现这个问题

	}
	//2、开启consumer 获得consumer的channel
	msgs, err := rabchannel.Consume(
		"filestore-channel-ceph",
		"gua",
		false, //是否自动应答      ***********不开启这个  使用手动回复 可靠性强  如果生产端没有受到ack 那么会再次发送一遍消息 这两次消息会堆积在队列中 下一次消息传送时会再次发送（堆积的以及这次的消息）
		false, //排他
		false, //不能将同一个conn中发送的消息传递
		false, //是否阻塞
		nil,   //其他参数
	)
	if err != nil {
		fmt.Println("rabconn consumer err:", err)
		return err
	} else {
		//错误处理 只不过提前在web端控制台设置好就不会出现这个问题

	}
	//3、启动携程处理 将msg交给处理函数 设置一个channel 使这个函数阻塞
	done := make(chan int, 1)
	go func() {
		for d := range msgs {
			//业务处理逻辑
			log.Printf("recive msg:%s\n", d.Body)
			err = callback(d.Body)
			//因为已经在生产端确认了消息到达队列 所以在这里只用ack nack reject 告知队列相关操作就行了
			if err != nil {
				//错误处理 //TODO :死信队列 也可以重新加入队列
				d.Reject(false) //false 丢弃 true 重新加入队列

			} else {
				//d.Ack(true) //false 丢弃 true 重新加入队列
				rabchannel.Ack(d.DeliveryTag, false) //tag用于标记是那一条消息
			}
		}

	}()
	<-done
	return nil
}

func MpRabConsumer(callback func(message []byte) error) error {
	//1、定义channel exchange 绑定 如果在控制台设置好 那么就不用
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
		return err
	} else {
		//错误处理 只不过提前在web端控制台设置好就不会出现这个问题

	}
	_, err = rabchannel.QueueDeclare(
		"filestore-channel-Mpceph",
		//是否持久化
		false,
		//是否自动删除
		false,
		//是否具有排他性 只能有一个消费者监听
		false,
		//是否堵塞
		false,
		nil, //额外属性  在这里是绑定死信机
	)
	if err != nil {
		fmt.Println("rabconn queue err:", err)
		return err
	} else {
		//错误处理 只不过提前在web端控制台设置好就不会出现这个问题

	}
	err = rabchannel.QueueBind(
		"filestore-channel-Mpceph",
		"Mpceph",
		"filestore-ceph",
		false,
		nil,
	)
	if err != nil {
		fmt.Println("rabconn bind err:", err)
		return err
	} else {
		//错误处理 只不过提前在web端控制台设置好就不会出现这个问题

	}
	//2、开启consumer 获得consumer的channel
	msgs, err := rabchannel.Consume(
		"filestore-channel-ceph",
		"gua",
		false, //是否自动应答      ***********不开启这个  使用手动回复 可靠性强  如果生产端没有受到ack 那么会再次发送一遍消息 这两次消息会堆积在队列中 下一次消息传送时会再次发送（堆积的以及这次的消息）
		false, //排他
		false, //不能将同一个conn中发送的消息传递
		false, //是否阻塞
		nil,   //其他参数
	)
	if err != nil {
		fmt.Println("rabconn consumer err:", err)
		return err
	} else {
		//错误处理 只不过提前在web端控制台设置好就不会出现这个问题

	}
	//3、启动携程处理 将msg交给处理函数 设置一个channel 使这个函数阻塞
	done := make(chan int, 1)
	go func() {
		for d := range msgs {
			//业务处理逻辑
			log.Printf("recive msg:%s\n", d.Body)
			err = callback(d.Body)
			//因为已经在生产端确认了消息到达队列 所以在这里只用ack nack reject 告知队列相关操作就行了
			if err != nil {
				//错误处理 //TODO :死信队列 也可以重新加入队列
				d.Reject(false) //false 丢弃 true 重新加入队列

			} else {
				//d.Ack(true) //false 丢弃 true 重新加入队列
				rabchannel.Ack(d.DeliveryTag, false) //tag用于标记是那一条消息
			}
		}

	}()
	<-done
	return nil
}

func DLXConsumer(callback func(message []byte) error) error {
	//1、定义channel exchange 绑定 如果在控制台设置好 那么就不用

	//2、开启consumer 获得consumer的channel
	msgs, err := rabchannel.Consume(
		"sixin_queue",
		"dlx",
		false, //是否自动应答      ***********不开启这个  使用手动回复 可靠性强  如果生产端没有受到ack 那么会再次发送一遍消息 这两次消息会堆积在队列中 下一次消息传送时会再次发送（堆积的以及这次的消息）
		false, //排他
		false, //不能将同一个conn中发送的消息传递
		false, //是否阻塞
		nil,   //其他参数
	)
	if err != nil {
		fmt.Println("rabconn consumer err:", err)
		return err
	} else {
		//错误处理 只不过提前在web端控制台设置好就不会出现这个问题

	}
	rabchannel.Qos(10, -1, false)
	//3、启动携程处理 将msg交给处理函数 设置一个channel 使这个函数阻塞
	done := make(chan int, 1)
	go func() {
		for d := range msgs {
			//业务处理逻辑
			log.Printf("recive msg:%s\n", d.Body)
			err = callback(d.Body)
			//因为已经在生产端确认了消息到达队列 所以在这里只用ack nack reject 告知队列相关操作就行了
			if err != nil {
				//错误处理 //TODO :死信队列 也可以重新加入队列
				d.Reject(true) //false 丢弃 true 重新加入队列

			} else {
				//d.Ack(true) //false 丢弃 true 重新加入队列
				rabchannel.Ack(d.DeliveryTag, false) //tag用于标记是那一条消息
			}
		}

	}()
	<-done
	return nil
}

//timeout+死信实现延时队列
func RedisDLXConsumer(callback func(message []byte) error) error {
	//1、定义channel exchange 绑定 如果在控制台设置好 那么就不用

	//2、开启consumer 获得consumer的channel
	msgs, err := rabchannel.Consume(
		"sixin_redis_queue",
		"dlx_redis",
		false, //是否自动应答      ***********不开启这个  使用手动回复 可靠性强  如果生产端没有受到ack 那么会再次发送一遍消息 这两次消息会堆积在队列中 下一次消息传送时会再次发送（堆积的以及这次的消息）
		false, //排他
		false, //不能将同一个conn中发送的消息传递
		false, //是否阻塞
		nil,   //其他参数
	)
	if err != nil {
		fmt.Println("rabconn consumer err:", err)
		return err
	} else {
		//错误处理 只不过提前在web端控制台设置好就不会出现这个问题

	}
	rabchannel.Qos(10, -1, false)
	//3、启动携程处理 将msg交给处理函数 设置一个channel 使这个函数阻塞
	done := make(chan int, 1)
	go func() {
		for d := range msgs {
			//业务处理逻辑
			log.Printf("recive msg:%s\n", d.Body)
			err = callback(d.Body)
			//因为已经在生产端确认了消息到达队列 所以在这里只用ack nack reject 告知队列相关操作就行了
			if err != nil {
				//错误处理 //TODO :死信队列 也可以重新加入队列
				d.Reject(true) //false 丢弃 true 重新加入队列

			} else {
				//d.Ack(true) //false 丢弃 true 重新加入队列
				rabchannel.Ack(d.DeliveryTag, false) //tag用于标记是那一条消息
			}
		}

	}()
	<-done
	return nil
}

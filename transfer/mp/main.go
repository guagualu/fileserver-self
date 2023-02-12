package main

import (
	"filestore/mq"
	"filestore/transfer/program"
	"fmt"
)

func main() {
	//1、启动rabblitmq的consumer
	var callback func(message []byte) error
	callback = program.CephMultiprogram
	err := mq.MpRabConsumer(callback)
	if err != nil {
		fmt.Println("消费端启动失败")
	}
}

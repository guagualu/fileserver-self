package main

import (
	"encoding/json"
	"filestore/db"
	rdlayer "filestore/db/redis"
	"filestore/mq"
	"fmt"
)

func DlxPragram(message []byte) error {
	//1、将message 从json转为golang类型 struct
	fileuserinfo := db.FileUserInfo{}
	err := json.Unmarshal(message, &fileuserinfo)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//2、redis删除
	rconn := rdlayer.RedisPool().Get()
	_, err = rconn.Do("HDEL", fileuserinfo.Username+"_file")
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
func main() {
	//1、启动rabblitmq的consumer
	var callback func(message []byte) error
	callback = DlxPragram
	err := mq.DLXConsumer(callback)
	if err != nil {
		fmt.Println("消费端启动失败")
	}
}

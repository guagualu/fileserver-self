package program

import (
	"encoding/json"
	"filestore/db"
	"filestore/mq"
	"filestore/store"
	"fmt"
	"io"
	"os"
)

func Cephprogram(message []byte) error {
	//1、将message 从json转为golang类型 struct
	fileinfo := mq.MqFileInfo{}
	err := json.Unmarshal(message, &fileinfo)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//2、找到file的临时存储地址，读取文件
	file, err := os.Open(fileinfo.CurLocateAt)
	if err != nil {
		fmt.Println(err)
		return err
	}
	filebyte, err := io.ReadAll(file)
	//3、获取ceph连接 并且将文件存储进去
	// bucket := store.GetCephBucket("filestoreself")
	path := "ceph/" + fileinfo.FileHash
	err = store.PutObject("ilestoreself", path, filebyte)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//4、ceph存入成功后，更新file表的locateAt
	err = db.UpdateFileLocateAt(fileinfo.FileHash, path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func CephMultiprogram(message []byte) error {
	//1、将message 从json转为golang类型 struct
	fileinfo := mq.MqFileInfo{}
	err := json.Unmarshal(message, &fileinfo)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//2、找到file的临时存储地址，读取文件
	file, err := os.Open(fileinfo.CurLocateAt)
	if err != nil {
		fmt.Println(err)
		return err
	}
	filebyte, err := io.ReadAll(file)
	//3、获取ceph连接 并且将文件存储进去
	// bucket := store.GetCephBucket("filestoreself")
	path := "ceph/" + fileinfo.FileHash
	err = store.MpPutObject("ilestoreself", path, filebyte)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//4、ceph存入成功后，更新file表的locateAt
	err = db.UpdateFileLocateAt(fileinfo.FileHash, path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

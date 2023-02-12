package store

import (
	"bytes"
	"fmt"

	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

var cephConn *s3.S3 //单例

// GetCephConnection : 获取ceph连接
func GetCephConnection() *s3.S3 {
	if cephConn != nil {
		return cephConn
	}
	// 1. 初始化ceph的一些信息

	auth := aws.Auth{ //权限相关 用户
		AccessKey: "MF5RJLR4QS9Y3LJT9786",
		SecretKey: "g31YU45q3T3CZ4IeHVwUlXUNetmqGPrbfrzfIpMw",
	}

	curRegion := aws.Region{
		Name:                 "default",                 //此区域的规范名称
		EC2Endpoint:          "http://172.21.0.15:7480", //服务器地址 docker ceph-rgw 对外开放的端口
		S3Endpoint:           "http://172.21.0.15:7480",
		S3BucketEndpoint:     "",    //no need
		S3LocationConstraint: false, //无区域限制
		S3LowercaseBucket:    false, //大小写不限制
		Sign:                 aws.SignV2,
	}

	// 2. 创建S3类型的连接
	return s3.New(auth, curRegion)
}

// GetCephBucket : 获取指定的bucket对象
func GetCephBucket(bucket string) *s3.Bucket {
	conn := GetCephConnection()
	return conn.Bucket(bucket)
}

// PutObject : 上传文件到ceph集群
func PutObject(bucket string, path string, data []byte) error {
	return GetCephBucket(bucket).Put(path, data, "octet-stream", s3.PublicRead)
}

//分块上传
func MpPutObject(bucket string, path string, data []byte) error {
	//1、使用bytes构建 bytereader
	breader := bytes.NewReader(data)

	//2、initmulti 获取 mutil结构
	m, err := GetCephBucket(bucket).InitMulti(path, "octet-stream", s3.PublicRead)
	if err != nil {
		fmt.Println("multi err:", err)
		return err
	}
	//3、使用mutil.putall 分块上传 获取每块part
	part, err := m.PutAll(breader, 5*1024)
	if err != nil {
		fmt.Println("putall err:", err)
		return err
	}
	//4、multi complete 组装到最终对象
	err = m.Complete(part)
	if err != nil {
		fmt.Println("putall err:", err)
		return err
	}
	return nil
}

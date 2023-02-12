package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"filestore/service/dbproxy/mapper"
	"filestore/service/dbproxy/orm"
	pb "filestore/service/dbproxy/proto"
)

type DBProxyStruct struct {
	pb.UnimplementedDBProxyServiceServer
}

func (d *DBProxyStruct) ExecAction(ctx context.Context, req *pb.ExecReq) (*pb.ExecResp, error) {
	//1、创建pb.resp 和作为data的resList
	resp := new(pb.ExecResp)
	resList := make([]orm.ExecRes, 0)
	//2、range req的ackton
	for k, singleAction := range req.Action {
		//3、处理singleaction的params todo 看 使用json转为 空接口数组  错误这个下标的code和suc等为错 data为nil
		params := []interface{}{}
		dec := json.NewDecoder(bytes.NewReader(singleAction.Params))
		dec.UseNumber() //防止转为float

		//将里面的内容decode进params里
		err := dec.Decode(&params)
		if err != nil {
			resList[k].Suc = false
			continue
		}
		//将 数字类型的param转为数字类型
		for _, v := range params {
			if _, ok := v.(json.Number); ok {
				params[k], _ = v.(json.Number).Int64()
			}
		}
		//4、执行函数 如果错误 错误处理 如果成功 将结果转为orm.execres类型 然后json化 变为次list下标的data
		rvlue, err := mapper.FuncCall(singleAction.Name, params...)
		if err != nil {
			resList[k].Suc = false
			continue
		}
		data, ok := (*rvlue)[0].Interface().(orm.ExecRes)
		if ok != true {
			resList[k].Suc = false
			continue
		}
		resList[k] = data
	}
	//5、将reslist json化
	data, err := json.Marshal(resList)
	if err != nil {
		return nil, err
	}
	resp.Data = data
	return resp, nil
}

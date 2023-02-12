package mapper

import (
	"errors"
	orm "filestore/service/dbproxy/orm"
	"reflect"
)

//动态代理db的持久层
var funcs = map[string]interface{}{
	"/file/GetFileInfo":        orm.GetFileInfo,
	"/file/GetFileInfoList":    orm.GetFileUserInfoList,
	"/file/UpdateFileLocation": orm.UpdateFileLocateAt,
	"/file/DeleteFile":         orm.DeletefileInfo,
	"/file/InsertFile":         orm.InsertFileInfo,

	"/fileuser/InsertFileUserInfo":  orm.InsertFileUserInfo,
	"/fileuser/UpdatefileUserInfo":  orm.UpdatefileUserInfo,
	"/fileuser/GetFileUserInfo":     orm.GetFileUserInfo,
	"/fileuser/GetFileUserInfoList": orm.GetFileUserInfoList,

	"/user/UserSignup": orm.SignupUserinfo,
	"/user/UserSignin": orm.SigninUserinfo,
	// "/user/GetUserInfo": ,
	// "/user/UserExist":   orm.UserExist,

	// "/ufile/OnUserFileUploadFinished": orm.OnUserFileUploadFinished,
	// "/ufile/QueryUserFileMetas":       orm.QueryUserFileMetas,
	// "/ufile/DeleteUserFile":           orm.DeleteUserFile,
	// "/ufile/RenameFileName":           orm.RenameFileName,
	// "/ufile/QueryUserFileMeta":        orm.QueryUserFileMeta,
}

func FuncCall(name string, params ...interface{}) (*[]reflect.Value, error) {
	//1、查找函数名字是否在funcs map中 不再报错
	f, ok := funcs[name]
	if !ok {
		return nil, errors.New("not in map")
	}

	//2、使用typeof将空接口转为type
	ftype := reflect.TypeOf(f)
	//3、判断 params的个数是否等于 函数的参数个数 调用value.innums 不对报错 昂type用new转为alue
	if ftype.NumIn() != len(params) {
		return nil, errors.New("params nums err")
	}
	fvalue := reflect.New(ftype)
	//4、调用value.call 返回
	param := make([]reflect.Value, 0)
	for k := range params {
		param = append(param, reflect.ValueOf(params[k]))
	}
	res := fvalue.Call(param)
	return &res, nil
}

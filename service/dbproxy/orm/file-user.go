package orm

import (
	dblay "filestore/service/dbproxy/conn"
)

func InsertFileUserInfo(username, filehash string) *ExecRes {
	stmt, err := dblay.DB().Prepare("insert into `file_user`(username,filehash,status) values(?,?,1)")
	execres := &ExecRes{}
	if err != nil {
		execres = ExecResFailed(execres)
		return execres
	}
	defer stmt.Close()
	res, err := stmt.Exec(username, filehash)
	if err != nil {
		execres = ExecResFailed(execres)
		return execres
	}
	if n, err := res.RowsAffected(); n >= 0 && err == nil {
		execres = ExecResFailed(execres)
		return execres
	}
	execres = ExecResSuc(execres, nil)
	return execres
}

//删除 改status
func UpdatefileUserInfo(username, filehash string) *ExecRes {
	stmt, err := dblay.DB().Prepare("update `file_user` set status=0 where username=? and filehash=?")
	execres := &ExecRes{}
	if err != nil {
		execres = ExecResFailed(execres)
		return execres
	}
	defer stmt.Close()
	res, err := stmt.Exec(username, filehash)
	if err != nil {
		execres = ExecResFailed(execres)
		return execres
	}
	if n, err := res.RowsAffected(); n >= 0 && err == nil {
		execres = ExecResFailed(execres)
		return execres
	}
	execres = ExecResSuc(execres, nil)
	return execres
}

//找到某个user是否有这个文件
func GetFileUserInfo(username, filehash string) (int, error) {
	stmt, err := dblay.DB().Prepare("select filehash from `file` where username=? and filehash=? and status =1")
	if err != nil {
		return -1, err
	}
	defer stmt.Close()
	res, err := stmt.Query(username)
	if err != nil {
		return -1, err
	}

	sum := 0
	for res.Next() {
		sum++
	}
	return sum, nil

}

//找到某个user的所有文件hash  有分页
func GetFileUserInfoList(username string, page, pagesize int) ([]string, error) {
	stmt, err := dblay.DB().Prepare("select filehash from `file` where username=? and status =1  limit ?,?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	res, err := stmt.Query(username, (page-1)*pagesize, pagesize)
	if err != nil {
		return nil, err
	}

	fileinfo := make([]string, 0)

	for res.Next() {
		var tmp string
		res.Scan(&tmp)
		fileinfo = append(fileinfo, tmp)
	}
	return fileinfo, nil

}

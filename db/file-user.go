package db

import dblay "filestore/db/mysql"

type FileUserInfo struct {
	Filename string `json:"filename"`
	Username string `json:"username"`
	Filehash string `json:"filehash"`
}

func InsertFileUserInfo(username, filehash string) error {
	stmt, err := dblay.DB().Prepare("insert into `file_user`(username,filehash,status) values(?,?,1)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(username, filehash)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); n >= 0 && err == nil {
		return nil
	}
	return err
}

//删除 改status
func UpdatefileUserInfo(username, filehash string) error {
	stmt, err := dblay.DB().Prepare("update `file_user` set status=0 where username=? and filehash=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(username, filehash)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); n >= 0 && err == nil {
		return nil
	}
	return err
}

//文件重命名
func UpdatefileName(filename, username, filehash string) error {
	stmt, err := dblay.DB().Prepare("update `file_user` set filename=? where username=? and filehash=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(filename, username, filehash)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); n >= 0 && err == nil {
		return nil
	}
	return err
}

//找到某个user的文件
func GetFileUserInfo(username, filehash string) (int, error) {
	stmt, err := dblay.DB().Prepare("select filehash from `file_user` where username=? and filehash=? and status =1")
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
func GetFileUserInfoList(username string, page, pagesize int) (*[]FileUserInfo, error) {
	stmt, err := dblay.DB().Prepare("select filehash,filename from `file_user` where username=? and status =1  limit ?,?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	res, err := stmt.Query(username, (page-1)*pagesize, pagesize)
	if err != nil {
		return nil, err
	}

	fileinfo := make([]FileUserInfo, 0)

	for res.Next() {
		var tmp FileUserInfo
		res.Scan(&tmp.Filename, &tmp.Filehash)
		fileinfo = append(fileinfo, tmp)
	}
	return &fileinfo, nil

}

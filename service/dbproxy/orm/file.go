package orm

import (
	dblay "filestore/service/dbproxy/conn"
)

type FileInfo struct {
	Filehash string
	Filename string
	LocateAt string //存在哪
}

func InsertFileInfo(fileinfo FileInfo) error {
	stmt, err := dblay.DB().Prepare("insert into `file`(filehash,filename,locateat) values(?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(fileinfo.Filehash, fileinfo.Filename, fileinfo.LocateAt)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); n >= 0 && err == nil {
		return nil
	}
	return err
}

func UpdateFileLocateAt(filehash, LocateAt string) error {
	stmt, err := dblay.DB().Prepare("update `file` set `locateat`=? where locateat=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(filehash, LocateAt)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); n >= 0 && err == nil {
		return nil
	}
	return err
}

func DeletefileInfo(fileinfo FileInfo) error {
	stmt, err := dblay.DB().Prepare("delete from `file` where filehash=? ")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(fileinfo.Filehash)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); n >= 0 && err == nil {
		return nil
	}
	return err
}

func GetFileInfo(filehash string) (*FileInfo, error) {
	stmt, err := dblay.DB().Prepare("select * from `file` where filehash=?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	res, err := stmt.Query(filehash)
	if err != nil {
		return nil, err
	}
	res.Next()
	fileinfo := &FileInfo{}
	err = res.Scan(fileinfo)
	if err != nil {
		return nil, err
	}
	return fileinfo, nil

}

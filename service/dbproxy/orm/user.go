package orm

import (
	mysql "filestore/service/dbproxy/conn"
	"fmt"
)

func SignupUserinfo(username, password string) error {
	stmt, err := mysql.DB().Prepare(fmt.Sprintf("insert into `user`(`username`,`password`) values(?,?)"))
	if err != nil {
		fmt.Println("signup gg err:", err)
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(username, password)
	if err != nil {
		fmt.Println("signup gg err:", err)
		return err
	}
	if n, _ := res.RowsAffected(); n <= 0 {
		fmt.Println("signup gg err:", err)
		return err
	}
	return nil

}

func SigninUserinfo(username, password string) error {
	stmt, err := mysql.DB().Prepare("select `username`,`password` from `user` where username =? and password =?")
	if err != nil {
		fmt.Println("signup gg err:", err)
		return err
	}
	defer stmt.Close()
	res, err := stmt.Query(username, password)
	if err != nil || res == nil {
		fmt.Println("signinuserinfo err:", err)
		return err
	}
	res.Next() //从第0行next到第1行返回记录
	tmpusername, tmppwd := "", ""
	err = res.Scan(&tmpusername, &tmppwd)
	if tmpusername == username && tmppwd == password && err != nil {
		return nil
	}
	return err

}

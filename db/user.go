package db

import (
	mysql "filestore/db/mysql"
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

type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

//查询用户信息
func GetUserInfo(username string) (*User, error) {
	user := User{}
	stmt, err := mysql.DB().Prepare(
		"select `username` from `user` where `username`=? limit 1",
	)
	if err != nil {
		fmt.Println("GetUserInfo err:", err.Error())
		return nil, err
	}
	defer stmt.Close()

	// 执行查询的操作
	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt) //第一行赋值给变量
	if err != nil {
		return &user, err
	}
	return &user, nil

}

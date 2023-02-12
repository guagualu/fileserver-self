package main

import (
	"filestore/service/apigw/route"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/user/signup", route.UserSignup)
	http.HandleFunc("/user/signin", route.UserSignin)
	http.HandleFunc("/user/userinfo", route.UserInfoHandler)
	http.HandleFunc("/user/getfilemeta", route.GetFileMetaHandler)
	http.HandleFunc("/user/filequery", route.FileQueryHandler)
	http.HandleFunc("/user/filerename", route.FileRenameHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("listen server err:", err)
	}
}

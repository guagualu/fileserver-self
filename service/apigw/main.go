package main

import (
	"filestore/service/apigw/route"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/user/signup", route.UserSignup)
	http.HandleFunc("/user/signin", route.UserSignin)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("listen server err:", err)
	}
}

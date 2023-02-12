package main

import (
	"filestore/service/upload/route"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/file/upload", route.FileUpload)
	http.HandleFunc("/user/mpuploadinit", route.Mpuploadinit)
	http.HandleFunc("/user/mpupload", route.Mpupload)
	http.HandleFunc("/user/findprogress", route.FindProgress)
	http.HandleFunc("/user/rempupload", route.ReMpUpload)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("listen server err:", err)
	}
}

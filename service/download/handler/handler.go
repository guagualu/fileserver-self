package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	mdb "filestore/db"
	resp "filestore/pkg/response"
	"filestore/store"
)

func DownloadURL(w http.ResponseWriter, r *http.Request) {
	//1、从前端获取字段 filehash
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	username := r.Form.Get("username")
	token := r.Form.Get("token")
	//2、从mysql的fileinfo是否是存在 从mysql获取存储位置 如果是ceph就走ceph的url 本地就走本地 如果是oss走oss 现在是只有ceph和本地
	n, err := mdb.GetFileUserInfo(username, filehash)
	if err != nil || n <= 0 {
		fmt.Println("get fileinfo err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	info, err := mdb.GetFileInfo(filehash)
	if err != nil {
		fmt.Println("get fileinfo err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	path := info.LocateAt
	var url string = ""
	if strings.HasPrefix(path, "ceph") || strings.HasPrefix(path, "./tmp") {

		url = fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s&locate=%s",
			r.Host, filehash, username, token)
	}

	//3、返回前端
	w.Write(resp.NewRespone(0, "success", url).ToJson())
}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	//1、从前端获取字段 filehash 位置标识
	r.ParseForm()
	filehash := r.Form.Get("filehash")

	//从mysql获取fileinfp 如果是ceph就走ceph的url 本地就走本地 如果是oss走oss 现在是只有ceph和本地

	info, err := mdb.GetFileInfo(filehash)
	if err != nil {
		fmt.Println("get fileinfo err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	path := info.LocateAt

	if strings.HasPrefix(path, "ceph") {
		bucket := store.GetCephBucket("filestoreself")
		data, err := bucket.Get(path)
		if err != nil {
			fmt.Println("get fileinfo ceph err:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octect-stream") //加一个header 方便浏览器识别 浏览器可以自动做下载
		w.Header().Set("Content-Description", "attachment;filename=\""+info.Filename+"\"")
		w.Write(resp.NewRespone(0, "success", data).ToJson())
	} else if strings.HasPrefix(path, "./tmp") {
		file, err := os.Open(path)
		if err != nil {
			fmt.Println(" file open err:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		data, err := io.ReadAll(file)
		if err != nil {
			fmt.Println(" file open err:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octect-stream") //加一个header 方便浏览器识别 浏览器可以自动做下载
		w.Header().Set("Content-Description", "attachment;filename=\""+info.Filename+"\"")
		w.Write(resp.NewRespone(0, "success", data).ToJson())
	}

	//3、返回前端
}

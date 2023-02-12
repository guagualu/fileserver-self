package route

import (
	"encoding/json"
	db "filestore/db"
	rdlayer "filestore/db/redis"
	"filestore/mq"
	"filestore/pkg"
	"filestore/pkg/jwt"
	resp "filestore/pkg/response"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

var salt = "filestore" //盐值 用于加密

func UserSignup(w http.ResponseWriter, r *http.Request) {
	//1、获取客户端字段 并进行有效性验证
	r.ParseForm()
	username := r.Form.Get("username")
	pwd := r.Form.Get("password")

	//todo 有效性验证
	//2、密码加密
	password := pkg.Sha1([]byte(pwd + salt))
	//3、进行db操作 insert操作
	err := db.SignupUserinfo(username, password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//4、跳转到登陆页面
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("注册成功"))
	http.Redirect(w, r, "http://127.0.0.1:8080/signin", http.StatusOK)
	return
}

func UserSignin(w http.ResponseWriter, r *http.Request) {
	//1、获取客户端字段
	r.ParseForm()
	username := r.Form.Get("username")
	pwd := r.Form.Get("password")
	//2、数据库获取验证
	password := pkg.Sha1([]byte(pwd + salt))
	err := db.SigninUserinfo(username, password)
	if err != nil {
		fmt.Println("usersignin err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//3、生成token
	token, err := jwt.GenerateToken(username, password)
	if err != nil {
		fmt.Println("token generate err:=", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//4、上传token 和成功信息(可以前端实现)
	res := resp.NewRespone(0, "succes", token)
	w.Write(res.ToJson())

}

//查询用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	//1、解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")

	//token验证已经放在了拦截器里面
	// token := r.Form.Get("token")
	// //2、验证token是否有效
	// isValidToken := IsTokenValid(token)
	// if !isValidToken {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	//3、查询用户信息  先查缓存 如果没有 在查mysql 并对redis作缓存
	rdconn := rdlayer.RedisPool().Get()
	rdres, err := rdconn.Do("GET", "user_"+username)
	if err != nil {
		user, err := db.GetUserInfo(username)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		//4、组装并且响应用户数据

		w.Write(resp.NewRespone(0, "OK", user).ToJson())
		msg, err := json.Marshal(user)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_, err = rdconn.Do("SET", "user_"+username, string(msg))
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		return
	}
	b, err := redis.Bytes(rdres, err)
	if err != nil {
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}
	user := db.User{}
	json.Unmarshal(b, &user)
	w.Write(resp.NewRespone(0, "OK", user).ToJson())

}

//查询文件元数据
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //将前端传过来的form 进行parse 解析到了r.Form

	filehash := r.Form["filehash"][0] //假设前端传来的字段名未filehash
	// fMeta := meta.GetFileMeta(filehash)

	//1、查询文件信息  先查缓存 如果没有 在查mysql 并对redis作缓存
	//2、做防止缓存穿透的场景 对 redis和mysql都没查到的情况 设置很短时间的空缓存 并对空缓存也做判断视为err 做单独的逻辑
	rdconn := rdlayer.RedisPool().Get()
	filebytes, err := redis.Bytes(rdconn.Do("GET", "file_"+filehash))
	if err != nil {
		fMeta, err := db.GetFileInfo(filehash)
		if err != nil {
			rdconn.Do("SET", "file_"+filehash, "EX", "3") ///超时3秒 防止一瞬间的攻击
			log.Fatal("文件元信息获取失败:", err)
			w.WriteHeader(http.StatusInternalServerError) //返回的头 的状态码
			return
		}
		//转为json格式并且上传前端
		filejson, err := json.Marshal(fMeta)
		if err != nil {
			log.Fatal("文件元信息转换失败:", err)
			w.WriteHeader(http.StatusInternalServerError) //返回的头 的状态码
			return
		}
		w.WriteHeader(200)
		w.Write(filejson) //返回前端
		return
	} else if filebytes == nil {
		log.Fatal("文件元信息获取失败:", err)
		w.WriteHeader(http.StatusInternalServerError) //返回的头 的状态码
		return
	}
	file := db.FileInfo{}
	json.Unmarshal(filebytes, &file)
	w.WriteHeader(200)
	w.Write(resp.NewRespone(0, "OK", file).ToJson()) //返回前端
	return

}

// FileQueryHandler : 查询批量的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	page, _ := strconv.Atoi(r.Form.Get("page"))
	username := r.Form.Get("username")
	//fileMetas, _ := meta.GetLastFileMetasDB(limitCnt)
	//1、查询redis 缓存 如果有 在返回数组中跳过（page-1）*limit 返回后续的limit个 如果没有 mysql逻辑 并缓存到redis
	rconn := rdlayer.RedisPool().Get()
	filenums, err := redis.Values(rconn.Do("HGETALL", username+"_file"))
	if err != nil {
		userFiles, err := db.GetFileUserInfoList(username, page, limitCnt)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for _, v := range *userFiles {
			jsoninfo, _ := json.Marshal(v)
			_, err := rconn.Do("HSET", username+"_file", v.Filehash, string(jsoninfo))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		data, err := json.Marshal(userFiles) //将结构体转为json形式的[]byte
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(data)
	}
	now := 0
	userFiles := make([]string, 0)
	for k, v := range filenums {
		if k < (page-1)*limitCnt {
			continue
		}
		if now >= limitCnt {
			break
		}
		userFiles = append(userFiles, v.(string))
	}
	w.Write(resp.NewRespone(0, "OK", userFiles).ToJson())

}

//user文件重命名
func FileRenameHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //将前端传过来的form 进行parse 解析到了r.Form

	filehash := r.Form.Get("filehash") //假设前端传来的字段名未filehash
	username := r.Form.Get("username") //假设前端传来的字段名未filehash
	filename := r.Form.Get("filename")
	// fMeta := meta.GetFileMeta(filehash)
	//1、删除redis 缓存
	rdconn := rdlayer.RedisPool().Get()
	_, err := rdconn.Do("HDEL", username+"_file", filehash)
	if err != nil {
		log.Fatal("文件重命名失败:", err)
		w.WriteHeader(http.StatusInternalServerError) //返回的头 的状态码
		return
	}

	//2、更新mysql
	err = db.UpdatefileName(filename, username, filehash)
	if err != nil {
		log.Fatal("文件重命名失败:", err)
		w.WriteHeader(http.StatusInternalServerError) //返回的头 的状态码
		return
	}
	userfileinfo := db.FileUserInfo{Filename: filename, Username: username, Filehash: filehash}
	info, _ := json.Marshal(userfileinfo)
	//3、异步延迟几秒 再删除redis
	mq.Redispublish("redis", string(info))
	w.WriteHeader(200)
	w.Write(resp.NewRespone(0, "OK", nil).ToJson()) //返回前端
	return

}

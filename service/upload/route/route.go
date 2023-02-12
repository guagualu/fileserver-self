package route

import (
	"bytes"
	"encoding/json"
	db "filestore/db"
	rpool "filestore/db/redis"
	"filestore/mq"
	"filestore/pkg"
	resp "filestore/pkg/response"
	"filestore/service/upload/config"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

type MpFileInfo struct {
	UploadId   string
	Filehash   string
	Filename   string
	Username   string
	ChunkSize  int
	ChunkCount int
}

func FileUpload(w http.ResponseWriter, r *http.Request) {
	//1、从前端获取文件，并暂时存在本地（会定时删除）
	r.ParseForm()
	username := r.Form.Get("username")
	file, fhead, err := r.FormFile("file")
	filename := fhead.Filename
	if err != nil {
		fmt.Println("file get err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = os.MkdirAll("./tmp", os.ModePerm)
	//2、获取file的hash值
	filebuffer := bytes.NewBuffer(nil)
	if _, err = io.Copy(filebuffer, file); err != nil {
		fmt.Println("file copy err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	filehash := pkg.Sha1(filebuffer.Bytes())
	filedst, err := os.Create("./tmp" + filehash)
	if _, err = io.Copy(filebuffer, file); err != nil {
		fmt.Println("file store err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(filedst, file)
	if err != nil {
		fmt.Println("file store err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//3、将fileinfo存入mysql ,以及file——user表
	fileinfo := db.FileInfo{

		Filehash: filehash,
		Filename: filename,
		LocateAt: "./tmp" + filehash,
	}
	err = db.InsertFileInfo(fileinfo)
	if err != nil {
		fmt.Println("file store err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = db.InsertFileUserInfo(username, filehash)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//4、异步将file转存到ceph 并修改文件的存储地址 todo
	mqfileinfo := mq.MqFileInfo{
		FileHash:    filehash,
		FileName:    filename,
		CurLocateAt: fileinfo.LocateAt,
	}
	msg, err := json.Marshal(mqfileinfo)
	if err != nil {
		fmt.Println("file store err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	mq.Rabpublish("ceph", string(msg))
	//5、返回成功给前端
	w.Write(resp.NewRespone(0, "succes", nil).ToJson())
}

func Mpuploadinit(w http.ResponseWriter, r *http.Request) {
	//1、从客户端获取字段 filesize 等等
	r.ParseForm()
	// filename:=r.Form.Get("filename")
	username := r.Form.Get("username")
	filename := r.Form.Get("filename")
	filehash := r.Form.Get("filehash")
	filesizetmp := r.Form.Get("filesize")
	filesize, err := strconv.Atoi(filesizetmp)
	if err != nil {
		fmt.Println("strconv err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//2、根据config 进行 计算 分多少块
	mpfileInfo := MpFileInfo{
		UploadId:   username + filename + filehash + time.Now().String(), //设置一个uploadid 规则
		Username:   username,
		Filename:   filename,
		Filehash:   filehash,
		ChunkCount: filesize / config.ChunkSize,
		ChunkSize:  config.ChunkSize,
	}
	//3、使用redis记录 使用map记录 filehash uploadid（相当于主键 可以自己定义规则 主要是用来唯一确定本次上传） todo
	redispool := rpool.RedisPool()
	redispool.Get().Do("HSET", "MP_"+mpfileInfo.UploadId, "CHUNKCOUNT", mpfileInfo.ChunkCount)
	// redispool.Get().Do("HSET","MP_"+mpfileInfo.UploadId,"",mpfileInfo.ChunkCount)
	//4、返回给前端
	w.Write(resp.NewRespone(0, "succes", mpfileInfo).ToJson())
	return

}

// 执行 linux shell command
func ExecLinuxShell(s string) (string, error) {
	//函数返回一个io.Writer类型的*Cmd
	cmd := exec.Command("/bin/bash", "-c", s)

	//通过bytes.Buffer将byte类型转化为string类型
	var result bytes.Buffer
	cmd.Stdout = &result

	//Run执行cmd包含的命令，并阻塞直至完成
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return result.String(), err
}

func Mpupload(w http.ResponseWriter, r *http.Request) {
	//1、从客户端获取消息 chunknum file uploadid redis 与mysql所需要的信息等
	r.ParseForm()
	file, fileheader, err := r.FormFile("file")
	if err != nil {
		fmt.Println("file get err:=", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	username := r.Form.Get("username")
	filename := fileheader.Filename
	uploadid := r.Form.Get("uploadid")
	chunknum := r.Form.Get("chunknum")
	chunkcount := r.Form.Get("chunkcount")
	filehash := r.Form.Get("filehash")
	//2、存入本地临时存储地点
	os.MkdirAll("./mpupload/"+filehash, 0777)
	filefinal, err := os.Create("./mpupload/" + filehash + "/" + chunknum)
	if err != nil {
		fmt.Println("file creat err:=", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(filefinal, file)
	if err != nil {
		fmt.Println("file copy err:=", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//3、redis 状态改变+1 如果chunknum==chunkcount 说明分块上传完成了 需要记录哪些块为1 总共需要多少块  status
	redispool := rpool.RedisPool()
	redispool.Get().Do("HSET", "MP_"+uploadid, "CHUNKINDEX_"+chunknum, 1)
	//redispool.Get().Do("HINCRBY", "MP_"+uploadid, "CHUNKINDEX_"+chunknum, 1)
	//4、如果上传完成了 将文件合并 并存入ceph中 rabbitmq的异步
	res, err := redis.Values(redispool.Get().Do("HGETALL", "MP_"+uploadid))
	if err != nil {
		fmt.Println("redis  err:=", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	flag := 0
	for i := 0; i < len(res); i = i + 2 {
		chunkindex := res[i]
		status := res[i+1]
		if tmp, ok := chunkindex.(string); ok && tmp == "CHUNKINDEX_"+chunkcount {
			if tmp, ok := status.(int); ok && tmp == 1 {
				flag = 1
			}

		}

	}
	//5、如果 上传完成 实现文件合并 异步到ceph 修改mysql中的file 和fileuser 合并todo
	if flag == 1 {
		// 使用linux命令构建  | 是管道命令 可以将左边的命令输出作为输入给右边的命令  xrags可以构建管道流 将左边输出流构建成管道流 | 给右边使用
		// .ls 不加任何参数,表示查询当前目录下的文件/文件夹
		// && 表示前一条命令执行成功时，才执行后一条命令 ，如 echo '1‘ && echo '2'
		cmd := fmt.Sprintf("cd %s && ls | sort -n | xargs cat > %s", "./mpupload/"+filehash, "./tmp/"+filehash)
		os.Create("./tmp/" + filehash)
		mergeRes, err := ExecLinuxShell(cmd)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Println(mergeRes)
	}
	fileinfo := db.FileInfo{
		Filehash: filehash,
		Filename: filename,
		LocateAt: "./tmp/" + filehash,
	}
	err = db.InsertFileInfo(fileinfo)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = db.InsertFileUserInfo(username, filehash)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//6、异步到ceph
	cephinfo := mq.MqFileInfo{
		FileHash:    filehash,
		FileName:    filename,
		CurLocateAt: "./tmp/" + filehash,
	}
	msg, err := json.Marshal(cephinfo)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	mq.Rabpublish("Mpceph", string(msg))

	//7、返回前端 如果上传完了 返回上传完成的 没上传完 但是这次成功 是另一种返回
	w.Write(resp.NewRespone(1, "分块上传完成", nil).ToJson())
}

//查询进度
func FindProgress(w http.ResponseWriter, r *http.Request) {
	//1、获取uploadid 、filename、username等redis上的所需要的
	uploadid := r.Form.Get("uploadid")
	chunkcount := r.Form.Get("chunkcount")
	count, err := strconv.Atoi(chunkcount)
	if err != nil {
		fmt.Println("redis  err:=", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//2、查询redis 如果状态位显示传完 那么就是返回完成 如果没有就返回num/count
	redispool := rpool.RedisPool()
	res, err := redis.Values(redispool.Get().Do("HGETALL", "MP_"+uploadid))
	if err != nil {
		fmt.Println("redis  err:=", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	sum := 0
	for i := 0; i < len(res); i = i + 2 {
		chunkindex := res[i]
		status := res[i+1]
		if tmp, ok := chunkindex.(string); ok && strings.HasPrefix(tmp, "CHUNKINDEX_") {
			if tmp, ok := status.(int); ok && tmp == 1 {
				sum += 1
			}
		}
	}
	progress := sum / count
	//3、返回前端
	w.Write(resp.NewRespone(0, "succes", progress).ToJson())
}

//重试
func ReMpUpload(w http.ResponseWriter, r *http.Request) {
	//1、获取uploadid 、filename、username等redis上的所需要的
	uploadid := r.Form.Get("uploadid")
	chunkcount := r.Form.Get("chunkcount")
	count, err := strconv.Atoi(chunkcount)
	if err != nil {
		fmt.Println("redis  err:=", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//2、查询redis哪些块是空的还没上传成功
	nofinished := make([]string, 0)
	tocompare := make(map[string]int)
	for i := 0; i < count; i++ {
		tocompare["CHUNKINDEX_"+strconv.Itoa(i)] = 0
	}

	redispool := rpool.RedisPool()
	res, err := redis.Values(redispool.Get().Do("HGETALL", "MP_"+uploadid))
	for i := 0; i < len(res); i = i + 2 {
		chunkindex := res[i]
		status := res[i+1]
		if tmp, ok := chunkindex.(string); ok && strings.HasPrefix(tmp, "CHUNKINDEX_") {
			if tmp1, ok := status.(int); ok && tmp1 == 1 {
				tocompare[tmp] = 1
			}
		}
	}
	//3、返回前端
	for k := range tocompare {
		if tocompare[k] != 1 {
			nofinished = append(nofinished, k)
		}
	}
	w.Write(resp.NewRespone(0, "succes", nofinished).ToJson())
	return
}

//尝试快传接口  这个也应该是前端在每次上传前调用
func TryFastUpload(w http.ResponseWriter, r *http.Request) {
	//1、从客户端获取字段 需要filehash和username
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	//2、在mysql中在file中查找file是否存在
	_, err := db.GetFileInfo(filehash)

	//3、如果file存在 将file写入file_user表中 如果不存在 返回前端不存在
	if err != nil {
		fmt.Println("redis  err:=", err)
		w.Write(resp.NewRespone(1, "Failed", nil).ToJson())
		return
	}
	err = db.InsertFileUserInfo(username, filehash)
	if err != nil {
		fmt.Println("redis  err:=", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//4、块传成功返回前端
	w.Write(resp.NewRespone(0, "Succes!", nil).ToJson())
}

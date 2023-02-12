package conn

//创建并发安全的 连接 db
import (
	mysql "database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" //驱动匿名打入 初始化 并将自己注册到database/sql里面区
)

var db *mysql.DB

func init() {
	var err error
	db, err = mysql.Open("mysql", "root:root1234@tcp(127.0.0.1:23307)/fileserver?charset=utf8")
	if err != nil {
		fmt.Println("db open err:", err)
		return
	}
	db.SetMaxOpenConns(10)
}
func DB() *mysql.DB {
	return db
}

package orm

//标准化返回的data各表的结构
type TabUser struct {
	Username string
	Password string
	Status   int
}

type TabFile struct {
	Filehash string
	Filename string
	LocateAt string
}

type TabUserFile struct {
	Username string
	Filehash string
	Status   int
}

type ExecRes struct {
	Suc     bool
	Code    int
	Message string
	Data    []byte
}

func ExecResFailed(execres *ExecRes) *ExecRes {
	execres.Suc = false
	execres.Code = -1
	execres.Message = "Failed"
	execres.Data = nil
	return execres
}

func ExecResSuc(execres *ExecRes, data []byte) *ExecRes {
	execres.Suc = true
	execres.Code = 0
	execres.Message = "Suc"
	execres.Data = data
	return execres
}

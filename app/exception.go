package app

type ApiException struct {
	Code   int    //http状态码
	Status int    //错误代码
	Msg    string //错误信息
}

func NewException(msg string, statusAndCod ...int) {
	code := 500
	status := 1
	if len(statusAndCod) >= 1 {
		status = statusAndCod[0]
	}
	if len(statusAndCod) >= 2 {
		code = statusAndCod[1]
	}

	ex := ApiException{
		Msg:    msg,
		Code:   code,
		Status: status,
	}
	panic(ex)
}

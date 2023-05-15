package Entity

type Response struct {
	Code int
	Data interface{}
}

// ReceiveCodeVo  当中的TestPointsVO 参数的格式应该为 ["1234","456578"]/*
type ReceiveCodeVo struct {
	Code         string `form:"code"`
	TestPointsVO string `form:"testPoints"`
	TestPoints   []string
	Language     string `form:"language"`
}

func (r Response) Fail(message string) (int, interface{}) {
	data := make(map[string]string)
	data["message"] = message
	response := Response{
		Code: 500,
		Data: data,
	}
	return 500, response
}
func (r Response) Success(message interface{}) (int, interface{}) {
	r.Code = 200
	r.Data = message
	return r.Code, r.Data
}

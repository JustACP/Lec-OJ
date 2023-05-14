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

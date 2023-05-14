package utils

import "oj/Entity"

func Fail(message string) (int, interface{}) {
	data := make(map[string]string)
	data["message"] = message
	response := Entity.Response{
		Code: 500,
		Data: data,
	}
	return 500, response
}

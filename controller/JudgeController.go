package controller

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/json"
	"oj/Entity"
	"oj/service"
	"oj/utils"
)

// content-type : application/x-www-form-urlencoded

func Controller(h *server.Hertz) {

	h.POST("/send", func(c context.Context, ctx *app.RequestContext) {
		var receive Entity.ReceiveCodeVo
		err := ctx.BindAndValidate(&receive)
		if err != nil {
			ctx.JSON(utils.Fail("消息绑定错误"))
		}
		err = json.Unmarshal([]byte(receive.TestPointsVO), &receive.TestPoints)
		if err != nil {
			ctx.JSON(utils.Fail("测试点转化错误"))
		}
		ctx.JSON(service.Judge(receive))
	})
	h.POST("/test", func(c context.Context, ctx *app.RequestContext) {
		var receive Entity.ReceiveCodeVo
		err := ctx.BindAndValidate(&receive)
		if err != nil {
			ctx.JSON(utils.Fail("消息绑定错误"))
		}
		err = json.Unmarshal([]byte(receive.TestPointsVO), &receive.TestPoints)
		fmt.Println(receive.TestPoints)
	})
}

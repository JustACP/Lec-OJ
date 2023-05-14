package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"oj/controller"
)

func main() {
	h := server.Default()
	controller.Controller(h)
	h.Spin()
}

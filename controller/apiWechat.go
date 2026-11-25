package controller

import (
	"github.com/kataras/iris/v12"
	"kandaoni.com/anqicms/provider"
)

func WechatApi(ctx iris.Context) {
	currentSite := provider.CurrentSite(ctx)
	wechatServer := currentSite.GetWechatServer(false)
	resp := wechatServer.VerifyURL(ctx.ResponseWriter(), ctx.Request())

	if ctx.Method() != "GET" {
		currentSite.ResponseWechatMsg(resp)
	}
}

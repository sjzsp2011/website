package controller

import (
	"log"

	"github.com/kataras/iris/v12"
	"github.com/medivhzhan/weapp/v3/server"
	"kandaoni.com/anqicms/model"
	"kandaoni.com/anqicms/provider"
)

func NotifyWeappMsg(ctx iris.Context) {
	currentSite := provider.CurrentSite(ctx)
	srv, err := server.NewServer(currentSite.PluginWeapp.AppID, currentSite.PluginWeapp.Token, currentSite.PluginWeapp.EncodingAESKey, "", "", true, nil)
	if err != nil {
		log.Printf("init server error: %s", err)
	}

	srv.OnSubscribeMsgPopup(func(popupEvent *server.SubscribeMsgPopupEvent) {
		SubscribeMsgPopup(ctx, popupEvent)
	})
	srv.OnSubscribeMsgChange(func(popupEvent *server.SubscribeMsgChangeEvent) {
		SubscribeMsgChange(ctx, popupEvent)
	})

	err = srv.Serve(ctx.ResponseWriter(), ctx.Request())
	if err != nil {
		log.Printf("serving error: %s", err)
	}
}

func SubscribeMsgPopup(ctx iris.Context, popupEvent *server.SubscribeMsgPopupEvent) {
	currentSite := provider.CurrentSite(ctx)
	for _, v := range popupEvent.SubscribeMsgPopupEvent {
		subscribedUser := model.SubscribedUser{
			Openid:     popupEvent.FromUserName,
			TemplateId: v.TemplateId,
		}
		if v.SubscribeStatusString == server.SubscribeResultAccept {
			//订阅
			currentSite.DB.Where("openid = ? AND template_id = ?", subscribedUser.Openid, subscribedUser.TemplateId).FirstOrCreate(&subscribedUser)
		} else {
			//拒绝
			currentSite.DB.Unscoped().Where("openid = ? AND template_id = ?", subscribedUser.Openid, subscribedUser.TemplateId).Delete(model.SubscribedUser{})
		}
	}
}

// SubscribeMsgChange 用户改变订阅消息事件
func SubscribeMsgChange(ctx iris.Context, popupEvent *server.SubscribeMsgChangeEvent) {
	currentSite := provider.CurrentSite(ctx)
	for _, v := range popupEvent.SubscribeMsgChangeEvent {
		subscribedUser := model.SubscribedUser{
			Openid:     popupEvent.FromUserName,
			TemplateId: v.TemplateId,
		}
		if v.SubscribeStatusString == server.SubscribeResultAccept {
			//订阅
			currentSite.DB.Where("openid = ? AND template_id = ?", subscribedUser.Openid, subscribedUser.TemplateId).FirstOrCreate(&subscribedUser)
		} else {
			//拒绝
			currentSite.DB.Unscoped().Where("openid = ? AND template_id = ?", subscribedUser.Openid, subscribedUser.TemplateId).Delete(model.SubscribedUser{})
		}
	}
}

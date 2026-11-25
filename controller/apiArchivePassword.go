package controller

import (
	"github.com/kataras/iris/v12"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/library"
	"kandaoni.com/anqicms/provider"
	"kandaoni.com/anqicms/request"
)

func ApiCheckArchivePassword(ctx iris.Context) {
	currentSite := provider.CurrentSite(ctx)
	var req request.ArchivePasswordRequest
	var err error
	if err = ctx.ReadJSON(&req); err != nil {
		req.Id = ctx.PostValueInt64Default("id", 0)
		req.Password = ctx.PostValueTrim("password")
		if req.Id == 0 || len(req.Password) == 0 {
			ctx.JSON(iris.Map{
				"code": config.StatusFailed,
				"msg":  err.Error(),
			})
			return
		}
	}

	archiveDetail, err := currentSite.GetArchiveById(req.Id)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}
	if len(archiveDetail.Password) == 0 || archiveDetail.Password == req.Password {
		var content string
		archiveData, err := currentSite.GetArchiveDataById(archiveDetail.Id)
		if err == nil {
			content = archiveData.Content
			// render
			if currentSite.Content.Editor == "markdown" {
				frontUrl := currentSite.System.BaseUrl
				if currentSite.System.FrontUrl != "" {
					frontUrl = currentSite.System.FrontUrl
				}
				content = library.MarkdownToHTML(archiveData.Content, frontUrl, currentSite.Content.FilterOutlink)
			}
		}
		detail := library.StructToMap(archiveDetail)
		detail["content"] = content
		detail["status"] = true
		delete(detail, "password")
		// extra
		archiveParams := currentSite.GetArchiveExtra(archiveDetail.ModuleId, archiveDetail.Id, true)
		for i := range archiveParams {
			if (archiveParams[i].Value == nil || archiveParams[i].Value == "") &&
				archiveParams[i].Type != config.CustomFieldTypeRadio &&
				archiveParams[i].Type != config.CustomFieldTypeCheckbox &&
				archiveParams[i].Type != config.CustomFieldTypeSelect {
				archiveParams[i].Value = archiveParams[i].Default
			}
			if archiveParams[i].FollowLevel && !archiveDetail.HasOrdered {
				delete(archiveParams, i)
			}
		}
		detail["extra"] = archiveParams
		ctx.JSON(iris.Map{
			"code": config.StatusOK,
			"msg":  "",
			"data": detail,
		})
		return
	}

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "",
		"data": iris.Map{
			"status":  false,
			"content": "",
		},
	})
}

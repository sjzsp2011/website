package manageController

import (
	"time"

	"github.com/kataras/iris/v12"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/model"
	"kandaoni.com/anqicms/provider"
	"kandaoni.com/anqicms/request"
)

// AnqiTranslateArticle 调用站点配置的翻译引擎，将指定文章（或草稿）的标题、描述、关键词和内容翻译为目标语言并保存。
func AnqiTranslateArticle(ctx iris.Context) {
	currentSite := provider.CurrentSite(ctx)
	var req request.TranslateArticleRequest
	var err error
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	isDraft := false
	archive, err := currentSite.GetArchiveById(req.Id)
	if err != nil {
		// 可能是 草稿
		archiveDraft, err := currentSite.GetArchiveDraftById(req.Id)
		if err != nil {
			ctx.JSON(iris.Map{
				"code": config.StatusFailed,
				"msg":  err.Error(),
			})
			return
		}
		isDraft = true
		archive = &archiveDraft.Archive
	}
	// 读取 data
	archiveData, err := currentSite.GetArchiveDataById(archive.Id)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}
	transReq := &provider.AnqiTranslateTextRequest{
		Text: []string{
			archive.Title,       // 0
			archive.Description, // 1
			archive.Keywords,    // 2
			archiveData.Content, // 3
		},
		Language:   currentSite.System.Language,
		ToLanguage: req.ToLanguage,
	}
	result, err := currentSite.AnqiTranslateString(transReq)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}
	// 更新文档
	archive.Title = result.Text[0]
	archive.Description = result.Text[1]
	archive.Keywords = result.Text[2]
	tx := currentSite.DB
	if isDraft {
		tx = tx.Model(&model.ArchiveDraft{})
	} else {
		tx = tx.Model(&model.Archive{})
	}
	tx.Where("id = ?", archive.Id).UpdateColumns(map[string]interface{}{
		"title":       archive.Title,
		"description": archive.Description,
		"keywords":    archive.Keywords,
	})
	// 再保存内容
	archiveData.Content = result.Text[3]
	currentSite.DB.Save(archiveData)

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  ctx.Tr("TranslationHasFinished"),
	})
}

// AnqiAiPseudoArticle 使用站点配置的写作模型改写指定文章或草稿。
//
// 参数说明：
//   - 请求体 "id": 文档 ID，支持正式文档或草稿。
func AnqiAiPseudoArticle(ctx iris.Context) {
	currentSite := provider.CurrentSite(ctx)
	var req request.AiPseudoArticleRequest
	var err error
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	isDraft := false
	archive, err := currentSite.GetArchiveById(req.Id)
	if err != nil {
		// 可能是 草稿
		archiveDraft, err := currentSite.GetArchiveDraftById(req.Id)
		if err != nil {
			ctx.JSON(iris.Map{
				"code": config.StatusFailed,
				"msg":  err.Error(),
			})
			return
		}
		isDraft = true
		archive = &archiveDraft.Archive
	}

	err = currentSite.AnqiAiPseudoArticle(archive, isDraft)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  ctx.Tr("AiPseudoOriginalHasBeenAddedToThePlan"),
	})
}

// AuthAiGenerateStream 使用站点配置的写作模型发起流式内容生成任务，返回流 ID 供后续读取。
//
// 参数说明：
//   - 请求体 "title": 关键词标题。
//   - 请求体 "demand": AI 生成的额外要求。
func AuthAiGenerateStream(ctx iris.Context) {
	currentSite := provider.CurrentSite(ctx)
	var req request.KeywordRequest
	var err error
	if err = ctx.ReadJSON(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	streamId, err := currentSite.AnqiAiGenerateStream(&req)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "",
		"data": streamId,
	})
}

// AuthAiGenerateStreamData 根据流 ID 轮询读取 AI 流式生成的当前内容，返回内容、提示信息及是否已完成。
//
// 参数说明：
//   - 查询参数 "stream_id": 流式生成任务 ID。
func AuthAiGenerateStreamData(ctx iris.Context) {
	currentSite := provider.CurrentSite(ctx)
	streamId := ctx.URLParam("stream_id")

	content, msg, finished := currentSite.AnqiLoadStreamData(streamId)

	if msg != "" {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  msg,
		})
		return
	}
	if finished {
		ctx.JSON(iris.Map{
			"code": config.StatusOK,
			"msg":  "finished",
			"data": content,
		})
		return
	}

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "",
		"data": content,
	})
}

// RestartAnqicms 重启 AnQiCMS 服务进程（通过重启信号通道，3 秒后由守护逻辑重新拉起）。
func RestartAnqicms(ctx iris.Context) {
	// first need to stop iris
	config.RestartChan <- config.RestartConfig{Code: 1, SiteId: 0}

	time.Sleep(3 * time.Second)

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  ctx.Tr("RestartSuccessfully"),
	})
}

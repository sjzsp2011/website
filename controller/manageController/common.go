package manageController

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	captcha "github.com/mojocn/base64Captcha"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/controller"
	"kandaoni.com/anqicms/model"
	"kandaoni.com/anqicms/provider"
)

// AdminFileServ 后台系统文件服务，根据请求路径输出 system 目录下的静态文件（支持目录默认 index.html）；同时根据 Site-Id 请求头或 site_id 参数为二级目录站点切换站点上下文。
func AdminFileServ(ctx iris.Context) {
	tmpSiteId := ctx.GetHeader("Site-Id")
	paramSiteId := ctx.URLParam("site_id")
	if tmpSiteId == "" && paramSiteId != "" {
		tmpSiteId = paramSiteId
	}
	if len(tmpSiteId) > 0 {
		siteId, _ := strconv.Atoi(tmpSiteId)
		if siteId > 0 {
			// 只有二级目录安装的站点允许这么操作
			website := provider.GetWebsite(uint(siteId))
			if website != nil && len(website.BaseURI) > 1 {
				ctx.Values().Set("siteId", uint(siteId))
			}
		}
	}

	uri := ctx.RequestPath(false)
	if uri != "/" {
		uriFile := config.ExecPath + strings.TrimLeft(uri, "/")
		// 防止路径遍历: 解析出真实路径后检查是否仍在允许的目录内
		uriFile = filepath.Clean(uriFile)
		basePath := config.ExecPath + "system"
		if !strings.HasPrefix(uriFile, basePath) {
			ctx.StatusCode(403)
			return
		}
		info, err := os.Stat(uriFile)
		if err == nil && !info.IsDir() {
			ctx.ServeFile(uriFile)
			return
		}

		if !strings.Contains(filepath.Base(uri), ".") {
			uriFile = strings.TrimRight(uriFile, "/") + "/index.html"
			_, err = os.Stat(uriFile)
			if err == nil {
				ctx.ServeFile(uriFile)
				return
			}
		}
	}

	ctx.Next()
}

// Version 获取当前系统版本号及是否为试用版。
func Version(ctx iris.Context) {
	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "",
		"data": iris.Map{
			"version": config.Version,
			"trial":   config.Trial,
		},
	})
}

// GetStatisticsSummary 获取统计摘要数据，exact 为 true 时进行精确统计。
//
// 参数说明：
//   - 查询参数 "exact": 是否精确统计，默认为 false。
func GetStatisticsSummary(ctx iris.Context) {
	currentSite := provider.CurrentSite(ctx)
	exact := ctx.URLParamBoolDefault("exact", false)
	result := currentSite.GetStatisticsSummary(exact)

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "",
		"data": result,
	})
}

// GetStatisticsDashboard 获取统计面板数据。
func GetStatisticsDashboard(ctx iris.Context) {
	currentSite := provider.CurrentSite(ctx)
	var result = iris.Map{}
	// 登录信息
	var loginLogs []model.AdminLoginLog
	currentSite.DB.Order("id desc").Limit(2).Find(&loginLogs)
	if len(loginLogs) == 2 {
		result["last_login"] = iris.Map{
			"created_time": loginLogs[1].CreatedTime,
			"ip":           loginLogs[1].Ip,
		}
	}
	if len(loginLogs) > 0 {
		result["now_login"] = iris.Map{
			"created_time": loginLogs[0].CreatedTime,
			"ip":           loginLogs[0].Ip,
		}
	}
	// 配置信息
	result["system"] = currentSite.System

	result["version"] = config.Version
	result["trial"] = config.Trial
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	result["memory_usage"] = ms.Alloc

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "",
		"data": result,
	})
}

// GenerateCaptcha 生成验证码
func GenerateCaptcha(ctx iris.Context) {
	currentSite := provider.CurrentSite(ctx)
	safeSetting := currentSite.Safe
	// 验证码是一次性凭据，绝不能被任何一层缓存回放：
	// 命中缓存会让前端拿到已作废的 captcha_id，表现为"点了图片不换"或验证码永远校验失败。
	ctx.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	ctx.Header("Pragma", "no-cache")                       // HTTP 1.0 兼容
	ctx.Header("Expires", "Thu, 01 Jan 1970 00:00:00 GMT") // 过期时间设为过去
	if safeSetting == nil || safeSetting.AdminCaptchaOff == 1 {
		ctx.JSON(iris.Map{
			"code": config.StatusOK,
			"msg":  "",
			"data": iris.Map{
				"captcha_off": true,
			},
		})
		return
	}

	var driver = controller.NewDriver().ConvertFonts()
	c := captcha.NewCaptcha(driver, controller.Store)
	id, content, answer := c.Driver.GenerateIdQuestionAnswer()
	item, _ := c.Driver.DrawCaptcha(content)
	c.Store.Set(id, answer)

	bs64 := item.EncodeB64string()

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "",
		"data": iris.Map{
			"captcha_off": false,
			"captcha_id":  id,
			"captcha":     bs64,
		},
	})
}

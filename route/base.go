package route

import (
	"github.com/kataras/iris/v12"
	"kandaoni.com/anqicms/controller"
	"kandaoni.com/anqicms/controller/graphql"
	"kandaoni.com/anqicms/middleware"
)

func Register(app *iris.Application) {
	//注册macros
	//设置错误
	app.Use(controller.Inspect)
	app.OnErrorCode(iris.StatusNotFound, controller.NotFound)
	app.OnErrorCode(iris.StatusInternalServerError, controller.InternalServerError)
	app.Use(controller.CheckTemplateType)
	app.Use(controller.Common)

	//由于使用了自定义路由，它不能同时解析两条到一起，因此这里不能启用fileserver，需要用nginx设置，有没研究出方法了再改进
	//app.HandleDir("/", fmt.Sprintf("%spublic", config.ExecPath)
	app.Get("/{path:path}", middleware.Check301, middleware.HandlerTimeout, controller.ReRouteContext)

	app.Get("/install", controller.Install)
	app.Post("/install", controller.InstallForm)
	app.HandleMany(iris.MethodPost, "/comment/publish /{base:string}/comment/publish", controller.CommentPublish)
	app.HandleMany(iris.MethodGet, "/comment/{id:uint} /{base:string}/comment/{id:uint}", middleware.HandlerTimeout, controller.CommentList)

	app.HandleMany(iris.MethodGet, "/guestbook.html /{base:string}/guestbook.html", middleware.HandlerTimeout, controller.GuestbookPage)
	app.HandleMany(iris.MethodPost, "/guestbook.html /{base:string}/guestbook.html", controller.GuestbookForm)

	// 内容导入API

	app.HandleMany(iris.MethodPost, "/attachment/upload /{base:string}/attachment/upload", controller.VerifyApiToken, controller.AttachmentUpload)
	app.HandleMany("GET POST", "/api/import/archive /{base:string}/api/import/archive", controller.VerifyApiToken, controller.ApiImportArchive)
	app.HandleMany("GET POST", "/api/import/categories /{base:string}/api/import/categories", controller.VerifyApiToken, controller.ApiImportGetCategories)
	app.HandleMany(iris.MethodPost, "/api/import/sitemap /{base:string}/api/import/sitemap", controller.VerifyApiToken, controller.ApiImportMakeSitemap)

	api := app.Party("/api")
	{
		// 注册GraphQL API v2路由
		graphql.RegisterRoutes(api)
		api.Get("/captcha", controller.GenerateCaptcha)
		// WeChat official account api
		api.Get("/wechat", controller.WechatApi)
		api.Post("/wechat", controller.WechatApi)
		api.Get("/log", controller.ApiLogStatistic)

		// 友链API
		api.Post("/friendlink/create", controller.VerifyApiLinkToken, controller.ApiImportCreateFriendLink)
		api.Post("/friendlink/delete", controller.VerifyApiLinkToken, controller.ApiImportDeleteFriendLink)
		api.Get("/friendlink/check", controller.VerifyApiLinkToken, controller.ApiImportCheckFriendLink)
		api.Post("/friendlink/check", controller.VerifyApiLinkToken, controller.ApiImportCheckFriendLink)
		api.Post("/archive/password/check", controller.ApiCheckArchivePassword)

		// common api
		api.Get("/archive/detail", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiArchiveDetail)
		api.Get("/archive/filters", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiArchiveFilters)
		api.Get("/archive/list", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiArchiveList)
		api.Get("/archive/params", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiArchiveParams)
		api.Get("/category/detail", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiCategoryDetail)
		api.Get("/category/list", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiCategoryList)
		api.Get("/module/detail", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiModuleDetail)
		api.Get("/module/list", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiModuleList)
		api.Get("/comment/list", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiCommentList)
		api.Get("/setting/contact", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiContact)
		api.Get("/setting/system", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiSystem)
		api.Get("/setting/diy", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiDiyField)
		api.Get("/setting/index", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiIndexTdk)
		api.Get("/languages", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiLanguages)
		api.Get("/guestbook/fields", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiGuestbook)
		api.Get("/friendlink/list", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiLinkList)
		api.Get("/nav/list", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiNavList)
		api.Get("/archive/next", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiNextArchive)
		api.Get("/archive/prev", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiPrevArchive)
		api.Get("/page/detail", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiPageDetail)
		api.Get("/page/list", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiPageList)
		api.Get("/tag/detail", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiTagDetail)
		api.Get("/tag/data/list", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiTagDataList)
		api.Get("/tag/list", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiTagList)
		api.Get("/banner/list", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiBannerList)
		api.Get("/place/detail", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiPlaceDetail)
		api.Get("/place/list", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiPlaceList)
		api.Get("/metadata", controller.CheckApiOpen, middleware.HandlerTimeout, controller.ApiMetadata)
		api.Post("/attachment/upload", controller.CheckApiOpen, controller.ApiAttachmentUpload)
		api.Post("/comment/publish", controller.CheckApiOpen, controller.ApiCommentPublish)
		api.Post("/guestbook.html", controller.CheckApiOpen, controller.ApiGuestbookForm)
		api.Post("/subscription", controller.CheckApiOpen, controller.ApiAddSubscriber)
	}

	notify := app.Party("/notify")
	{
		notify.Get("/weapp/msg", controller.NotifyWeappMsg)
	}

	//后台管理路由相关
	manageRoute(app)

	// MCP 对外端点：按域名解析站点 + Bearer token 鉴权
	app.Any("/api/mcp", middleware.McpTokenAuth, controller.McpStreamableHTTP)
}

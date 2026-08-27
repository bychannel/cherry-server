package web

import (
	checkCenter "github.com/bychannel/cherry-server/internal/component/check_center"
	"github.com/bychannel/cherry-server/internal/data"
	"github.com/bychannel/cherry-server/nodes/web/controller"
	"github.com/bychannel/cherry-server/nodes/web/sdk"
	"github.com/cherry-game/cherry"
	cherryCron "github.com/cherry-game/components/cron"
	cherryGin "github.com/cherry-game/components/gin"
	"github.com/gin-gonic/gin"
)

func Run(profileFilePath, nodeID string) {
	// 配置cherry引擎,加载profile配置文件
	app := cherry.Configure(profileFilePath, nodeID, false, cherry.Cluster)

	// 注册调度组件
	app.Register(cherryCron.New())

	// 注册检查中心服是否启动组件
	app.Register(checkCenter.New())

	// 注册数据配表组件
	app.Register(data.New())

	// 加载http server组件
	app.Register(httpServerComponent(app.Address()))

	// 加载sdk逻辑
	sdk.Init(app)

	// 启动cherry引擎
	app.Startup()
}

func httpServerComponent(addr string) *cherryGin.Component {
	gin.SetMode(gin.DebugMode)

	// new http server
	httpServer := cherryGin.NewHttp("http_server", addr)
	httpServer.Use(cherryGin.Cors())

	// http server使用gin组件搭建，这里增加一个RecoveryWithZap中间件
	httpServer.Use(cherryGin.RecoveryWithZap(true))

	// 注册 controller
	httpServer.Register(new(controller.LoginController))

	return httpServer
}

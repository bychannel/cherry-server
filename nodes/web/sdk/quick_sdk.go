package sdk

import (
	"github.com/bychannel/cherry-server/internal/code"
	"github.com/bychannel/cherry-server/internal/data"
	sessionKey "github.com/bychannel/cherry-server/internal/session_key"
	cerror "github.com/cherry-game/cherry/error"
	cherryHttp "github.com/cherry-game/cherry/extend/http"
	cstring "github.com/cherry-game/cherry/extend/string"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryGin "github.com/cherry-game/components/gin"
)

type (
	quickSdk struct {
	}
)

func (quickSdk) SdkId() int32 {
	return QuickSDK
}

func (quickSdk) Login(config *data.SdkRow, params Params, callback Callback) {
	token, found := params.GetString("token")
	if found == false || cstring.IsBlank(token) {
		err := cerror.Error("token is nil")
		callback(code.LoginError, nil, err)
		return
	}

	uid, found := params.GetString("uid")
	if found == false || cstring.IsBlank(uid) {
		err := cerror.Error("uid is nil")
		callback(code.LoginError, nil, err)
		return
	}

	rspBody, _, err := cherryHttp.GET(config.LoginURL(), map[string]string{
		"token":        token,
		"uid":          uid,
		"product_code": config.GetString("productCode"),
	})

	if err != nil || rspBody == nil {
		callback(code.LoginError, nil, err)
		return
	}

	var result = string(rspBody)
	if result != "1" {
		cherryLogger.Warnf("quick sdk login fail. rsp =%s", rspBody)
		callback(code.LoginError, nil, err)
		return
	}

	callback(code.OK, map[string]string{
		sessionKey.OpenID: uid, // 返回 quick的uid做为 open id
	})
}

func (s quickSdk) PayCallback(config *data.SdkRow, c *cherryGin.Context) {
	// TODO 这里实现quick sdk 支付回调的逻辑
	c.RenderHTML("FAIL")
}

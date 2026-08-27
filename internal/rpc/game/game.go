package rpcGame

import (
	"fmt"

	"github.com/bychannel/cherry-server/internal/pb"
	sessionKey "github.com/bychannel/cherry-server/internal/session_key"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
)

const (
	playerActor = "player"
)

const (
	sessionClose = "sessionClose"
)

// SessionClose 如果session已登录，则调用rpcGame.SessionClose() 告知游戏服
func SessionClose(app cfacade.IApplication, session *cproto.Session) {
	nodeID := session.GetString(sessionKey.ServerID)
	if nodeID == "" {
		clog.Warnf("Get server id fail. session = %s", session.Sid)
		return
	}

	targetPath := fmt.Sprintf("%s.%s.%s", nodeID, playerActor, session.Sid)
	app.ActorSystem().Call("", targetPath, sessionClose, &pb.Int64{
		Value: session.Uid,
	})

	// clog.Infof("send close session to game node. [node = %s, uid = %d]", nodeID, session.Uid)
}

package account

import (
	"strings"

	"github.com/bychannel/cherry-server/internal/code"
	"github.com/bychannel/cherry-server/internal/pb"
	"github.com/bychannel/cherry-server/nodes/center/db"
	cactor "github.com/cherry-game/cherry/net/actor"
)

type (
	ActorAccount struct {
		cactor.Base
	}
)

func (p *ActorAccount) AliasID() string {
	return "account"
}

// OnInit center为后端节点，不直接与客户端通信，所以了一些remote函数，供RPC调用
func (p *ActorAccount) OnInit() {
	p.Remote().Register("registerDevAccount", p.registerDevAccount)
	p.Remote().Register("getDevAccount", p.getDevAccount)
	p.Remote().Register("getUserId", p.getUserId)
}

// registerDevAccount 注册开发者帐号
func (p *ActorAccount) registerDevAccount(req *pb.DevRegister) int32 {
	accountName := req.AccountName
	password := req.Password

	if strings.TrimSpace(accountName) == "" || strings.TrimSpace(password) == "" {
		return code.LoginError
	}

	if len(accountName) < 3 || len(accountName) > 18 {
		return code.LoginError
	}

	if len(password) < 3 || len(password) > 18 {
		return code.LoginError
	}

	return db.DevAccountRegister(accountName, password, req.Ip)
}

// getDevAccount 根据帐号名获取开发者帐号表
func (p *ActorAccount) getDevAccount(req *pb.DevRegister) (*pb.Int64, int32) {
	accountName := req.AccountName
	password := req.Password

	devAccount, _ := db.DevAccountWithName(accountName)
	if devAccount == nil || devAccount.Password != password {
		return nil, code.AccountAuthFail
	}

	return &pb.Int64{Value: devAccount.AccountId}, code.OK
}

// getUserId 获取userId
func (p *ActorAccount) getUserId(req *pb.User) (*pb.Int64, int32) {
	userId, ok := db.BindUserId(req.SdkId, req.PackageId, req.OpenId)
	if userId == 0 || ok == false {
		return nil, code.AccountBindFail
	}

	return &pb.Int64{Value: userId}, code.OK
}

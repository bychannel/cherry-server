package main

import (
	"encoding/json"
	"fmt"
	"github.com/bychannel/cherry-server/tools/robot/base"
	"github.com/bychannel/cherry-server/tools/robot/conf"
	"math/rand"
	"sync"
	"time"

	"github.com/bychannel/cherry-server/internal/code"
	chttp "github.com/cherry-game/cherry/extend/http"
	clog "github.com/cherry-game/cherry/logger"
	pomeloClient "github.com/cherry-game/cherry/net/parser/pomelo/client"
)

var cfg *conf.Config

func main() {
	cfg = conf.LoadConfig("./conf/config.json")

	wg := sync.WaitGroup{}
	wg.Add(1)

	accounts := make(map[string]string)
	for i := 1; i <= int(cfg.MaxRobotNum); i++ {
		key := fmt.Sprintf("test%d", i)
		accounts[key] = key
	}

	// 第一步，先到web节点进行账号注册
	RegisterDevAccount(cfg.WebUrl, accounts)

	for userName, password := range accounts {
		time.Sleep(time.Duration(rand.Int31n(10)) * time.Millisecond)
		go RunRobot(cfg.WebUrl, cfg.Pid, userName, password, cfg.GateAddr, cfg.ServerId, cfg.PrintLog)
	}

	wg.Wait()
}

func RegisterDevAccount(webUrl string, accounts map[string]string) {
	requestURL := fmt.Sprintf("%s/register", webUrl)

	for key, val := range accounts {
		params := map[string]string{
			"account":  key,
			"password": val,
		}

		jsonBytes, _, err := chttp.GET(requestURL, params)
		if err != nil {
			clog.Warn(err)
			return
		}

		rsp := &code.Result{}
		err = json.Unmarshal(jsonBytes, rsp)
		if err != nil {
			clog.Warn(err)
			return
		}

		clog.Debugf("register account = %s, result = %+v", key, rsp)
	}
}

func RunRobot(webUrl, pid, userName, password, gateAddr string, serverId int32, printLog bool) *base.Robot {

	// 创建客户端
	cli := base.New(
		pomeloClient.New(
			pomeloClient.WithRequestTimeout(10*time.Second),
			pomeloClient.WithErrorBreak(true),
		),
	)
	cli.PrintLog = printLog

	// 登录获取token
	if err := cli.GetLoginToken(webUrl, pid, userName, password); err != nil {
		clog.Error(err)
		return nil
	}

	// 根据地址连接网关
	if err := cli.ConnectToTCP(gateAddr); err != nil {
		clog.Error(err)
		return nil
	}

	if cli.PrintLog {
		clog.Infof("tcp connect %s is ok", gateAddr)
	}

	// 随机休眠
	cli.RandSleep()

	// 用户登录到游戏节点
	err := cli.UserLogin(serverId)
	if err != nil {
		clog.Warn(err)
		return nil
	}

	if cli.PrintLog {
		clog.Infof("user login is ok. [user = %s, serverId = %d]", userName, serverId)
	}

	// cli.RandSleep()

	// 查看是否有角色
	err = cli.PlayerSelect()
	if err != nil {
		clog.Warn(err)
		return nil
	}

	// cli.RandSleep()

	// 创建角色
	err = cli.ActorCreate()
	if err != nil {
		clog.Warn(err)
		return nil
	}

	// cli.RandSleep()

	// 角色进入游戏
	err = cli.ActorEnter()
	if err != nil {
		clog.Warn(err)
		return nil
	}

	elapsedTime := cli.StartTime.NowDiffMillisecond()
	clog.Debugf("[%s] is enter to game. elapsed time:%dms", cli.TagName, elapsedTime)

	// cli.Disconnect()

	return cli
}

package conf

import (
	"encoding/json"
	clog "github.com/cherry-game/cherry/logger"
	"os"
)

// Config 机器人配置文件
type Config struct {
	MaxRobotNum int32  `json:"maxRobotNum"` // 运行x个机器人
	WebUrl      string `json:"webUrl"`      // web node 地址
	GateAddr    string `json:"gateAddr"`    // 网关地址(正式环境通过区服列表获取)
	ServerId    int32  `json:"serverId"`    // 测试的游戏服id
	Pid         string `json:"pid"`         // 测试的sdk包id
	PrintLog    bool   `json:"printLog"`    // 是否输出详细日志
}

// LoadConfig 加载配置文件，加载失败则 panic
func LoadConfig(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		panic("load config failed: " + err.Error())
	}

	tmpCfg := &Config{}
	if err = json.Unmarshal(data, tmpCfg); err != nil {
		panic("parse config failed: " + err.Error())
	}

	clog.Debugf("load config success. result = %+v", tmpCfg)
	return tmpCfg
}
